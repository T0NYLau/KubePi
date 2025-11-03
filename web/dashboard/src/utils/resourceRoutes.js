export const mixin = {
  methods: {
    toResource (type, namespace, name) {
      let routeUrl;
      // 获取正确的cluster名称，处理this.cluster可能是对象的情况
      let clusterName = '';
      if (this.clusterName) {
        clusterName = this.clusterName;
      } else if (this.cluster) {
        // 如果this.cluster是对象，提取其中的名称
        if (typeof this.cluster === 'object' && this.cluster !== null) {
          clusterName = this.cluster.name || this.cluster.clusterName || '';
        } else {
          clusterName = this.cluster;
        }
      }
      
      switch (type) {
        case "Pod":
          routeUrl = this.$router.resolve({
            name: "PodDetail",
            params: { namespace: namespace, name: name },
            query: { yamlShow: false, cluster: clusterName }
          })
          window.open(routeUrl.href, "_blank")
          break
        case "Service":
          routeUrl = this.$router.resolve({
            name: "ServiceDetail",
            params: { namespace: namespace, name: name },
            query: { yamlShow: false, cluster: clusterName }
          })
          window.open(routeUrl.href, "_blank")
          break
        case "Job":
          routeUrl = this.$router.resolve({
            name: "JobDetail",
            params: { namespace: namespace, name: name },
            query: { yamlShow: false, cluster: clusterName }
          })
          window.open(routeUrl.href, "_blank")
          break
        case "HorizontalPodAutoscaler":
          routeUrl = this.$router.resolve({
            name: "HPADetail",
            params: { namespace: namespace, name: name },
            query: { yamlShow: false, cluster: clusterName }
          })
          window.open(routeUrl.href, "_blank")
          break
        case "PersistentVolumeClaim":
          routeUrl = this.$router.resolve({
            name: "PersistentVolumeClaimDetail",
            params: { namespace: namespace, name: name },
            query: { yamlShow: false, cluster: clusterName }
          })
          window.open(routeUrl.href, "_blank")
          break
        case "Endpoints":
          routeUrl = this.$router.resolve({
            name: "EndpointDetail",
            params: { namespace: namespace, name: name },
            query: { yamlShow: false, cluster: clusterName }
          })
          window.open(routeUrl.href, "_blank")
          break
        case "PodDisruptionBudget":
          routeUrl = this.$router.resolve({
            name: "PDBDetail",
            params: { namespace: namespace, name: name },
            query: { yamlShow: false, cluster: clusterName }
          })
          window.open(routeUrl.href, "_blank")
          break
        default:
          break
      }
    }
  }
}

