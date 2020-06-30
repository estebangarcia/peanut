/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	// log is for logging in this package.
	peanutpipelinelog = logf.Log.WithName("peanutpipeline-resource")
	c                 client.Client
	ctx               = context.Background()
	group             = "pipeline.peanut.dev"
	kind              = "PeanutPipeline"
)

func (r *PeanutPipeline) SetupWebhookWithManager(mgr ctrl.Manager) error {
	c = mgr.GetClient()

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-pipeline-peanut-dev-v1alpha1-peanutpipeline,mutating=true,failurePolicy=fail,groups=pipeline.peanut.dev,resources=peanutpipelines,verbs=create;update,versions=v1alpha1,name=mpeanutpipeline.peanut.dev

var _ webhook.Defaulter = &PeanutPipeline{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *PeanutPipeline) Default() {
	peanutpipelinelog.Info("default", "name", r.Name)

}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-pipeline-peanut-dev-v1alpha1-peanutpipeline,mutating=false,failurePolicy=fail,groups=pipeline.peanut.dev,resources=peanutpipelines,versions=v1alpha1,name=vpeanutpipeline.peanut.dev

var _ webhook.Validator = &PeanutPipeline{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *PeanutPipeline) ValidateCreate() error {
	peanutpipelinelog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return r.validatePipeline()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *PeanutPipeline) ValidateUpdate(old runtime.Object) error {
	peanutpipelinelog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return r.validatePipeline()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *PeanutPipeline) ValidateDelete() error {
	peanutpipelinelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *PeanutPipeline) validatePipeline() error {
	var allErrs field.ErrorList

	for i, stage := range r.Spec.Stages {
		path := field.NewPath("spec").Child("stages").Index(i)

		for j, aux := range r.Spec.Stages {
			if stage.Name == aux.Name && i != j {
				allErrs = append(allErrs, field.Duplicate(path, stage.Name))
			}
		}

		namespace := types.NamespacedName{
			Namespace: r.Namespace,
			Name:      stage.Type,
		}

		if _, err := getStage(namespace); err != nil {
			allErrs = append(allErrs, field.NotFound(path, err.Error()))
		}
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{
		Group: group,
		Kind:  kind,
	}, r.Name, allErrs)
}

func getStage(key client.ObjectKey) (*PeanutStage, error) {
	peanutStage := &PeanutStage{}
	if err := c.Get(ctx, key, peanutStage); err != nil {
		return nil, err
	}

	return peanutStage, nil
}
