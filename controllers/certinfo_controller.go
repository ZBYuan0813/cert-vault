/*
Copyright 2020 ms.

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
	certv1 "cert-vault/api/v1"
	"cert-vault/pkg"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// CertInfoReconciler reconciles a CertInfo object
type CertInfoReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const myFinalizerName = "cert.vault.crd.finalizers"
// +kubebuilder:rbac:groups=cert.vault.com,resources=certinfoes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert.vault.com,resources=certinfoes/status,verbs=get;update;patch

func (r *CertInfoReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("certinfo", req.NamespacedName)

	// your logic here
	cert := &certv1.CertInfo{}
	secret := &corev1.Secret{}
	if err := r.Get(ctx, req.NamespacedName, cert); err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		return ctrl.Result{}, nil
	}

	if cert.ObjectMeta.DeletionTimestamp.IsZero(){
		log.Info("IsZero","IsZero",cert.ObjectMeta.DeletionTimestamp.IsZero())
		if !containsString(cert.ObjectMeta.Finalizers, myFinalizerName){
			//cert.ObjectMeta.Finalizers = append(cert.ObjectMeta.Finalizers,myFinalizerName)
			controllerutil.AddFinalizer(cert,myFinalizerName)
			if err := r.Update(context.Background(),cert); err != nil{
				return ctrl.Result{}, nil
			}
		}
		if err := r.getSecret(ctx,secret,cert); err != nil {
			log.Info("Create secret field","Error info",err)
			return ctrl.Result{},nil
		}
	} else {
			if containsString(cert.ObjectMeta.Finalizers,myFinalizerName){
				if err := r.reconcileDelete(ctx,cert); err != nil {
					return ctrl.Result{}, err
				}
				controllerutil.RemoveFinalizer(cert,myFinalizerName)
				//cert.ObjectMeta.Finalizers=removeString(cert.ObjectMeta.Finalizers,myFinalizerName)
				if err := r.Update(context.Background(), cert); err != nil {
					return ctrl.Result{}, err
				}
			}
	}
	return ctrl.Result{}, nil
}

func (r *CertInfoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&certv1.CertInfo{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// the func for delete ca from vault and crd
func (r *CertInfoReconciler) reconcileDelete(ctx context.Context,cert *certv1.CertInfo) error{
	logger := r.Log.WithValues("cluster", cert.Name, "namespace", cert.Namespace)
	logger.Info("Delete cert crd and Revoke Cert")
	secret := &corev1.Secret{}
	var req ctrl.Request
	req.Name = cert.Name
	req.Namespace = cert.Namespace
	if err := r.Client.Get(ctx,req.NamespacedName,secret); err !=nil{
		return err
	}
	client := pkg.CreateVaultConfig()
	serialNum := string(secret.Data["serial_number"])
	revoke := pkg.RevokeData{
		SerialNumber: serialNum,
	}
	logger.Info("Revoke Cert","serial_number",serialNum)
	revokeData,_ := json.Marshal(revoke)
	revokePath := "/v1/"+cert.Spec.Path+"/revoke"
	statusCode := pkg.RevokeCert(revokePath,revokeData,client)
	logger.Info("Revoke Cert response","StatusCode:",statusCode)
	return nil
}

// create role and cert from root ca use root token
func (r *CertInfoReconciler) generateCert(certInfo *certv1.CertInfo) (ca map[string][]byte){
	logger := r.Log.WithValues("cluster", certInfo.Name, "namespace", certInfo.Namespace)
	role := pkg.Role{
		RoleName:  certInfo.Spec.RoleName,
		RoleData: pkg.RoleData{
			Allowed_Domains: []string{"mskj.com"},
			Allow_subdomains: true,
			Allow_Any_Name: true,
			Organization: certInfo.Spec.Organization,
			Ou: certInfo.Spec.Ou,
			Max_TTL: certInfo.Spec.Max_TTL,
		},
	}

	cert := pkg.Cert{
		RoleName: certInfo.Spec.RoleName,
		CertData: pkg.CertData{
			CommonName: certInfo.Spec.CommonName,
		},
	}

	rolebody,_ := json.Marshal(role.RoleData)

	certbody, _ := json.Marshal(cert.CertData)

	client := pkg.CreateVaultConfig()
	res := pkg.CreateRole("/v1/"+certInfo.Spec.Path+"/roles/",role.RoleName,rolebody,client)
	logger.Info("Create role response","response",string(res))
	m := pkg.CreateCert("/v1/"+certInfo.Spec.Path+"/issue/",cert.RoleName,certbody,client)
	return m
}

// find secret whether exits
func (r *CertInfoReconciler) getSecret(ctx context.Context,secret *corev1.Secret,cert *certv1.CertInfo) error {
	log := r.Log.WithValues("Secret namespace",cert.Namespace)
	var req ctrl.Request
	req.Name = cert.Name
	req.Namespace = cert.Namespace
	if err := r.Client.Get(ctx,req.NamespacedName,secret); err != nil{
		log.Info("secret create secret","secret",req.NamespacedName)
		// 创建secret
		m := r.generateCert(cert)
		return r.createSecret(ctx,m,cert)
		}else {
			log.Info("Secret exits")
			//m := r.generateCert(cert)
			//secret.Data = m
			//log.Info("Update secret")
			//return r.updateSecret(ctx,secret)
		}
	log.Info("End secret")
	return nil
}

// set secret and add cert info to secret data
func (r *CertInfoReconciler)createSecret(ctx context.Context,m map[string][]byte,cert *certv1.CertInfo) error{
	sc := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: cert.Name,
			Namespace: cert.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cert.GetObjectMeta(),cert.GroupVersionKind()),
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Secret",
		},
		Data: m,
	}
	if err := r.Client.Create(ctx, sc); err != nil {
			return err
	}
	return nil
}


func (r *CertInfoReconciler)updateSecret(ctx context.Context,sc *corev1.Secret) error{
	if err := r.Client.Update(ctx, sc); err != nil {
		return err
	}
	return nil
}