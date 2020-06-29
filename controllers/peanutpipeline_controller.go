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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	peanut "github.com/estebangarcia/peanut/api/v1alpha1"
)

// PeanutPipelineReconciler reconciles a PeanutPipeline object
type PeanutPipelineReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=pipeline.peanut.dev,resources=peanutpipelines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pipeline.peanut.dev,resources=peanutpipelines/status,verbs=get;update;patch

func (r *PeanutPipelineReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("peanutpipeline", req.NamespacedName)

	var pipeline peanut.PeanutPipeline
	if err := r.Get(ctx, req.NamespacedName, &pipeline); err != nil {
		log.Error(err, "unable to fetch PeanutPipeline")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	pipeline.Status.Ack = "true"
	if err := r.Status().Update(ctx, &pipeline); err != nil {
		log.Error(err, "unable to update PeanutPipeline status")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *PeanutPipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&peanut.PeanutPipeline{}).
		Complete(r)
}
