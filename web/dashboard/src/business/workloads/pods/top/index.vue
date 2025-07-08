<template>
  <div>
    <el-alert v-if="showText" :title="$t('business.pod.metric_server_tip')" type="warning" />
    <br>
    <el-button style="margin-left: 20px" :disabled="sortType==='namespace'" icon="el-icon-sort-down" type="text" @click="setSort('namespace')">{{ $t('business.namespace.namespace') }} / {{ $t('business.workload.name') }}</el-button>
    <el-button icon="el-icon-sort-down" :disabled="sortType==='cpu'" type="text" @click="setSort('cpu')">CPU</el-button>
    <el-button icon="el-icon-sort-down" :disabled="sortType==='memory'" type="text" @click="setSort('memory')">{{ $t('business.workload.memory') }}</el-button>
    <el-button icon="el-icon-refresh" type="text" @click="refresh">{{ $t('commons.button.refresh') }}</el-button>

    <complex-table :data="data" v-loading="loading">
      <el-table-column :label="$t('business.namespace.namespace')" prop="metadata.namespace" min-width="60" show-overflow-tooltip fix>
        <template v-slot:default="{row}">
          {{ row.metadata.namespace }}
        </template>
      </el-table-column>
      <el-table-column :label="$t('commons.table.name')" prop="metadata.name" show-overflow-tooltip min-width="80">
        <template v-slot:default="{row}">
          <span class="span-link" @click="openDetail(row)">{{ row.metadata.name }}</span>
        </template>
      </el-table-column>
      <el-table-column :label="$t('business.workload.container_name')" show-overflow-tooltip min-width="80">
        <template v-slot:default="{row}">
          <div v-for="(c, index) in row.containers" :key="index">
            <div>
              <span>{{ c.name }}</span>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="CPU" min-width="35">
        <template v-slot:default="{row}">
          <div v-for="(c, index) in row.containers" :key="index">
            <div>
              <span>{{ c.usage.cpu | cpu }}</span>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column :label="$t('business.workload.memory')" min-width="35">
        <template v-slot:default="{row}">
          <div v-for="(c, index) in row.containers" :key="index">
            <div>
              <span>{{ c.usage.memory | mi-memory }} Mi</span>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column :label="$t('business.pod.time_stamp')" min-width="35" show-overflow-tooltip prop="metadata.timestamp" fix>
        <template v-slot:default="{row}">
          {{ row.timestamp | age }}
        </template>
      </el-table-column>
      <el-table-column :label="$t('business.pod.windows')" show-overflow-tooltip prop="window" min-width="35" fix>
        <template v-slot:default="{row}">
          {{ row.window }}
        </template>
      </el-table-column>
    </complex-table>

    <div class="dialog-footer" style="margin-top: 20px; text-align: right;">
      <el-button @click="$emit('close')">{{ $t("commons.button.close") }}</el-button>
    </div>
  </div>
</template>

<script>
import ComplexTable from "@/components/complex-table"
import { listPodMetrics } from "@/api/apis"

export default {
  name: "PodTop",
  components: { ComplexTable },
  data() {
    return {
      loading: false,
      showText: false,
      data: [],
      sortType: 'namespace',
    }
  },
  methods: {
    refresh() {
      this.listPodMetric();
    },
    setSort(type) {
      this.sortType = type;
      if (!this.data) return;
      let arr = [...this.data];
      if (type === 'namespace') {
        arr.sort((a, b) => {
          if (a.metadata.namespace < b.metadata.namespace) return -1;
          if (a.metadata.namespace > b.metadata.namespace) return 1;
          if (a.metadata.name < b.metadata.name) return -1;
          if (a.metadata.name > b.metadata.name) return 1;
          return 0;
        });
      } else if (type === 'cpu') {
        arr.sort((a, b) => (b.cpu || 0) - (a.cpu || 0));
      } else if (type === 'memory') {
        arr.sort((a, b) => (b.memory || 0) - (a.memory || 0));
      }
      this.data = arr;
    },
    listPodMetric() {
      this.loading = true;
      const namespace = sessionStorage.getItem("namespace");
      const cluster = this.$route.query.cluster;
      listPodMetrics(cluster, namespace)
        .then((res) => {
          // 为每个item补充cpu和memory字段
          this.data = (res.items || []).map(item => {
            let cpu = 0, memory = 0;
            if (item.containers && item.containers.length > 0) {
              for (const c of item.containers) {
                // 这里假设c.usage.cpu为字符串如"12345n"，c.usage.memory为"12345Ki"
                if (c.usage && c.usage.cpu) {
                  cpu += Number(String(c.usage.cpu).replace(/[^\d.]/g, ""));
                }
                if (c.usage && c.usage.memory) {
                  memory += Number(String(c.usage.memory).replace(/[^\d.]/g, ""));
                }
              }
            }
            // cpu单位转换（如有需要可调整）
            item.cpu = cpu;
            item.memory = memory;
            return item;
          });
          this.setSort(this.sortType); // 默认排序
          if (this.data.length === 0) {
            this.showText = true;
          }
        })
        .finally(() => {
          this.loading = false;
        });
    },
    openDetail(row) {
      this.$router.push({
        name: "PodDetail",
        params: { namespace: row.metadata.namespace, name: row.metadata.name },
        query: { yamlShow: false },
      })
    },
  },
  mounted() {
    this.sortType = 'namespace';
    this.listPodMetric();
  }
}
</script>

