package files

import (
	"context"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	secretstoreapi "github.com/crossplane-contrib/provider-gitlab/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/features"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errNotFile          = "managed resource is not a Gitlab file custom resource"
	errProjectIDMissing = "ProjectID is missing"
	errKubeUpdateFailed = "cannot update Gitlab file custom resource"
	errCreateFailed     = "cannot create Gitlab file"
	errUpdateFailed     = "cannot update Gitlab file"
	errDeleteFailed     = "cannot delete Gitlab file"
	errGetFailed        = "cannot retrieve Gitlab file with"
)

// SetupFile adds a controller that reconciles Files.
func SetupFile(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.FileKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), secretstoreapi.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: projects.NewFileClient}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.FileGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.File{}).
		Complete(r)
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) projects.FileClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.File)
	if !ok {
		return nil, errors.New(errNotFile)
	}
	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	kube   client.Client
	client projects.FileClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.File)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotFile)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(errProjectIDMissing)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	fileOptions := projects.GenerateGetFileOptions(&cr.Spec.ForProvider, e.kube, ctx)
	file, res, err := e.client.GetFile(
		*cr.Spec.ForProvider.ProjectID,
		*cr.Spec.ForProvider.FilePath,
		fileOptions,
		gitlab.WithContext(ctx))
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(projects.IsFileNotFound, err), errGetFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	projects.LateInitializeFile(&cr.Spec.ForProvider, file)

	cr.Status.AtProvider = projects.GenerateFileObservation(file)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        projects.IsFileUpToDate(&cr.Spec.ForProvider, file),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.File)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotFile)
	}

	cr.Status.SetConditions(xpv1.Creating())
	fileOptions := projects.GenerateCreateFileOptions(&cr.Spec.ForProvider, e.kube, ctx)

	file, _, err := e.client.CreateFile(*cr.Spec.ForProvider.ProjectID, *cr.Spec.ForProvider.FilePath, fileOptions, gitlab.WithContext(ctx))

	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, file.FilePath)
	return managed.ExternalCreation{}, errors.Wrap(err, errKubeUpdateFailed)

}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.File)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotFile)
	}

	updateOptions := projects.GenerateUpdateFileOptions(&cr.Spec.ForProvider, e.kube, ctx)
	_, _, err := e.client.UpdateFile(*cr.Spec.ForProvider.ProjectID, *cr.Spec.ForProvider.FilePath, updateOptions, gitlab.WithContext(ctx))

	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)

}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.File)
	if !ok {
		return errors.New(errNotFile)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	deleteOptions := projects.GenerateDeleteFileOptions(&cr.Spec.ForProvider, e.kube, ctx)
	_, err := e.client.DeleteFile(*cr.Spec.ForProvider.ProjectID, *cr.Spec.ForProvider.FilePath, deleteOptions, gitlab.WithContext(ctx))
	return errors.Wrap(err, errDeleteFailed)
}
