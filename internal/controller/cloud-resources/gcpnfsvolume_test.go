package cloudresources

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	skriprange "github.com/kyma-project/cloud-manager/pkg/skr/iprange"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Created SKR GcpNfsVolume is projected into KCP and it gets Ready condition and PV created", func() {

	Context("Given SKR Cluster", Ordered, func() {

		It("And Given SKR namespace exists", func() {
			Eventually(CreateNamespace).
				WithArguments(infra.Ctx(), infra.SKR().Client(), &corev1.Namespace{}).
				Should(Succeed())
		})

		skrIpRangeName := "gcp-iprange-1"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		kcpIpRangeName := "513f20b4-7b73-4246-9397-f8dd55344479"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}
		skrGcpNfsVolumeName := "gcp-nfs-volume-1"
		nfsIpAddress := "10.11.12.14"

		It("And Given SKR IpRange exists", func() {

			// tell skriprange reconciler to ignore this SKR IpRange
			skriprange.Ignore.AddName(skrIpRangeName)

			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
				).
				Should(Succeed())
		})

		It("And Given KCP IpRange exists", func() {

			// tell kcpiprange reconciler to ignore this KCP IpRange
			kcpiprange.Ignore.AddName(kcpIpRangeName)

			Eventually(CreateKcpIpRange).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName),
					WithNamespace(DefaultKcpNamespace),
					WithLabels(map[string]string{
						cloudcontrolv1beta1.LabelKymaName:        infra.SkrKymaRef().Name,
						cloudcontrolv1beta1.LabelRemoteName:      skrIpRangeName,
						cloudcontrolv1beta1.LabelRemoteNamespace: DefaultSkrNamespace,
					}),
				).
				Should(Succeed())
		})

		It("And KCP IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		It("And SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusCidr(skrIpRange.Spec.Cidr),
					WithSkrIpRangeStatusId(kcpIpRangeName),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		gcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}

		It("When GcpNfsVolume is created", func() {
			Eventually(CreateGcpNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
					WithName(skrGcpNfsVolumeName),
					WithGcpNfsVolumeIpRange(skrIpRange.Name),
				).
				Should(Succeed())
		})

		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}

		It("Then GcpNfsVolume gets ID from the KCP NfsInstance", func() {
			// load GcpNfsVolume to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpNfsVolume,
					NewObjActions(),
					AssertGcpNfsVolumeHasId(),
				).
				Should(Succeed())
		})

		It("Then Load KCP NfsInstance with the specified identifier", func() {
			// check KCP NfsInstance is created with name=gcpNfsVolume.ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpNfsInstance,
					NewObjActions(WithName(gcpNfsVolume.Status.Id)),
				).
				Should(Succeed())
		})

		It("And KCP NfsInstance should be properly created", func() {
			By("KCP NfsInstance should have label cloud-manager.kyma-project.io/kymaName")
			Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))

			By("KCP NfsInstance should have label cloud-manager.kyma-project.io/remoteName")
			Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(gcpNfsVolume.Name))

			By("KCP NfsInstance should have label cloud-manager.kyma-project.io/remoteNamespace")
			Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(gcpNfsVolume.Namespace))

			By("KCP NfsInstance should have spec.scope.name equal to SKR Cluster kyma name")
			Expect(kcpNfsInstance.Spec.Scope.Name).To(Equal(infra.SkrKymaRef().Name))

			By("KCP IpRange should have spec.remoteRef matching to to SKR IpRange")
			Expect(kcpNfsInstance.Spec.RemoteRef.Namespace).To(Equal(gcpNfsVolume.Namespace))
			Expect(kcpNfsInstance.Spec.RemoteRef.Name).To(Equal(gcpNfsVolume.Name))
		})

		It("When KCP NfsInstance is switched to Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
					WithConditions(KcpReadyCondition()),
					WithKcpNfsStatusState(cloudcontrolv1beta1.ReadyState),
					WithKcpNfsStatusHost(nfsIpAddress),
				).
				Should(Succeed())
		})

		It("Then SKR GcpNfsVolume will get to Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpNfsVolume,
					NewObjActions(),
				).
				Should(Succeed())
		})
	})

})
