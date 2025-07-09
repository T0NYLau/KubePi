export const mixin = {
  methods: {
    toResource (type, namespace, name) {
      let routeUrl;
      switch (type) {
        case "Pod":
          routeUrl = this.$router.resolve({
            name: "PodDetail",
            params: { namespace: namespace, name: name },
            query: { yamlShow: false, cluster: this.cluster }
          })
          window.open(routeUrl.href, "_blank")
          break
        case "Service":
          routeUrl = this.$router.resolve({
            name: "ServiceDetail",
            params: { namespace: namespace, name: name },
            query: { yamlShow: false, cluster: this.cluster }
          })
          window.open(routeUrl.href, "_blank")
          break
        case "Job":
          routeUrl = this.$router.resolve({
            name: "JobDetail",
            params: { namespace: namespace, name: name },
            query: { yamlShow: false, cluster: this.cluster }
          })
          window.open(routeUrl.href, "_blank")
          break
        case "HorizontalPodAutoscaler":
          routeUrl = this.$router.resolve({
            name: "HPADetail",
            params: { namespace: namespace, name: name },
            query: { yamlShow: false, cluster: this.cluster }
          })
          window.open(routeUrl.href, "_blank")
          break
        case "PersistentVolumeClaim":
          routeUrl = this.$router.resolve({
            name: "PersistentVolumeClaimDetail",
            params: { namespace: namespace, name: name },
            query: { yamlShow: false, cluster: this.cluster }
          })
          window.open(routeUrl.href, "_blank")
          break
        case "Endpoints":
          routeUrl = this.$router.resolve({
            name: "EndpointDetail",
            params: { namespace: namespace, name: name },
            query: { yamlShow: false, cluster: this.cluster }
          })
          window.open(routeUrl.href, "_blank")
          break
        case "PodDisruptionBudget":
          routeUrl = this.$router.resolve({
            name: "PDBDetail",
            params: { namespace: namespace, name: name },
            query: { yamlShow: false, cluster: this.cluster }
          })
          window.open(routeUrl.href, "_blank")
          break
        default:
          break
      }
    }
  }
}

