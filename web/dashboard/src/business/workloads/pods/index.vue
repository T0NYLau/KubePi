<template>
  <div>
  <layout-content header="Pods">
    <div style="float: left">
      <el-button type="primary" size="small" @click="onCreate"
                   v-has-permissions="{scope:'namespace',apiGroup:'',resource:'pods',verb:'create'}">
        YAML
      </el-button>
      <el-button type="primary" size="small" @click="onTop"
                  v-has-permissions="{scope:'namespace',apiGroup:'',resource:'pods',verb:'list'}">
        Top
      </el-button>
      <el-button type="primary" size="small" :disabled="selects.length===0" @click="onDelete()"
                  v-has-permissions="{scope:'namespace',apiGroup:'',resource:'pods',verb:'delete'}">
        {{ $t("commons.button.delete") }}
      </el-button>
      <el-button type="primary" size="small" :disabled="selects.length===0" @click="onForceDelete()"
                  v-has-permissions="{scope:'namespace',apiGroup:'',resource:'pods',verb:'delete'}">
        {{ $t("commons.button.delete_force") }}
      </el-button>
      <el-button type="primary" size="small"
                   @click="exportToXlsx()" icon="el-icon-download">
        {{ $t("commons.button.export") }}
      </el-button>

      <el-button type="primary" size="small"
                   @click="batchTerminal()"  :disabled="selects.length===0 || !checkExecPermissions()">
        {{ $t("commons.button.terminal") }}
      </el-button>

      <el-button type="primary" size="small"
                   @click="batchLogs()"  :disabled="selects.length===0 || !checkLogPermissions()">
        {{ $t("commons.button.logs") }}
      </el-button>
    </div>
    <complex-table :selects.sync="selects" :data="data" v-loading="loading" :pagination-config="paginationConfig"
                   :search-config="searchConfig" @search="search" :showFullTextSwitch="true" @update:isFullTextSearch="OnIsFullTextSearchChange" @sort-change="sortTableFun">
      <el-table-column type="selection" fix ></el-table-column>
      <el-table-column :label="$t('commons.table.name')" prop="name" min-width="80" show-overflow-tooltip fix sortable="name">
        <template v-slot:default="{row}">
          <span class="span-link" @click="openDetail(row)">
            {{ row.metadata.name }}
          </span>
        </template>
      </el-table-column>
      <el-table-column :label="$t('business.event.event')" min-width="30" prop="event_count" sortable="event_count">
        <template v-slot:default="{row}">
          <el-button 
            v-if="podEvents[`${row.metadata.namespace}-${row.metadata.name}`] && podEvents[`${row.metadata.namespace}-${row.metadata.name}`].length > 0"
            type="text" 
            @click="showPodEvents(row)"
            style="padding: 0; font-size: 16px;">
            <i class="el-icon-warning" style="color: #E6A23C;"></i>
          </el-button>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column :label="$t('business.namespace.namespace')" min-width="45" prop="namespace" sortable="namespace"/>
      <el-table-column :label="$t('commons.table.status')" min-width="30" sortable="status_phase" prop="status_phase">
        <template v-slot:default="{row}">
          <div v-if="row.status.phase ==='Running'">
            <i class="el-icon-check" />&nbsp; &nbsp; &nbsp;
            {{ $t("commons.status.Running") }}
          </div>
          <div v-if="row.status.phase ==='Failed'">
            <i class="el-icon-close" />&nbsp; &nbsp; &nbsp;
            {{ $t("commons.status.Failed") }}
          </div>
          <div v-if="row.status.phase ==='Pending'">
            <i class="el-icon-loading" />&nbsp; &nbsp; &nbsp;
            {{ $t("commons.status.Pending") }}
          </div>
          <div v-if="row.status.phase ==='Succeeded'">
            <i class="el-icon-finished" />&nbsp; &nbsp; &nbsp;
            {{ $t("commons.status.Succeeded") }}
          </div>
          <div v-if="row.status.phase ==='Unknown'">
            <i class="iconfont iconwenhao" />&nbsp; &nbsp; &nbsp;
            {{ $t("commons.status.Unknown") }}
          </div>
        </template>
      </el-table-column>
      <el-table-column :label="$t('commons.table.ready')" min-width="30" prop="pod_status" sortable="pod_status">
        <template v-slot:default="{row}">
          {{ getPodStatus(row) }}
        </template>
      </el-table-column>
      <el-table-column label="IP" min-width="40" prop="podIP" sortable="podIP"/>
      <el-table-column :label="$t('business.cluster.nodes')" min-width="45" show-overflow-tooltip prop="nodeName" sortable="nodeName"/>
      <el-table-column :label="$t('commons.table.created_time')" show-overflow-tooltip min-width="35" prop="creationTimestamp" fix sortable="creationTimestamp">
        <template v-slot:default="{row}">
          {{ row.metadata.creationTimestamp | age }}
        </template>
      </el-table-column>
      <el-table-column :label="$t('business.pod.restart_count')" show-overflow-tooltip min-width="35"  fix sortable="restart_count" prop="restart_count">
        <template v-slot:default="{row}">
          {{ getRestartTimes(row) }}
        </template>
      </el-table-column>
      <el-table-column width="90px" :label="$t('commons.table.action')">
        <template v-slot:default="{row}">
          <el-button circle @click="onEdit(row)" size="mini" icon="el-icon-edit"
                     v-has-permissions="{scope:'namespace',apiGroup:'',resource:'pods',verb:'update'}"/>
          <el-dropdown style="margin-left: 10px" @command="handleClick($event,row)" :hide-on-click="false">
            <el-button circle icon="el-icon-more" size="mini"/>
            <el-dropdown-menu slot="dropdown">
              <div v-if="row.containers.length > 1">
                <el-popover placement="left" trigger="hover">
                  <div v-for="c in row.containers" :key="c">
                    <p style="margin: 0">
                      <el-button :disabled="!checkExecPermissions()" @click="openTerminal(row, c)" type="text">{{ c }}</el-button>
                    </p>
                  </div>
                  <el-dropdown-item slot="reference" :disabled="!checkExecPermissions()" icon="el-icon-date" command="terminal">
                    {{ $t("commons.button.terminal") }}
                    <i class="el-icon-arrow-right"/>
                  </el-dropdown-item>
                </el-popover>
                <el-popover placement="left" trigger="hover">
                  <div v-for="c in row.containers" :key="c">
                    <p style="margin: 0">
                      <el-button :disabled="!checkLogPermissions()" @click="openTerminalLogs(row, c)" type="text">{{ c }}</el-button>
                    </p>
                  </div>
                  <el-dropdown-item slot="reference" :disabled="!checkLogPermissions()" icon="el-icon-notebook-2" command="logs">
                    {{ $t("commons.button.logs") }}
                    <i class="el-icon-arrow-right"/>
                  </el-dropdown-item>
                </el-popover>
                <el-popover placement="left" trigger="hover">
                  <div v-for="c in row.containers" :key="c">
                    <p style="margin: 0">
                      <el-button :disabled="!checkExecPermissions()" @click="openPodFiles(row, c)" type="text">{{ c }}</el-button>
                    </p>
                  </div>
                  <el-dropdown-item slot="reference" icon="el-icon-files" command="files">
                    {{ $t("business.pod.pod_file") }}
                    <i class="el-icon-arrow-right"/>
                  </el-dropdown-item>
                </el-popover>
              </div>
              <div v-if="row.containers.length == 1">
                <el-dropdown-item :disabled="!checkExecPermissions()" icon="iconfont iconline-terminalzhongduan" command="terminal">
                  {{ $t("commons.button.terminal") }}
                </el-dropdown-item>
                <el-dropdown-item :disabled="!checkLogPermissions()" icon="el-icon-tickets" command="logs">{{ $t("commons.button.logs") }}
                </el-dropdown-item>
                <el-dropdown-item :disabled="!checkExecPermissions()" icon="el-icon-files" command="files">{{ $t("business.pod.pod_file") }}
                </el-dropdown-item>
              </div>
              <el-dropdown-item icon="el-icon-download" command="download">{{ $t("commons.button.download_yaml") }}</el-dropdown-item>
              <el-dropdown-item icon="el-icon-delete" :disabled="!onCheckDeletePermissions()" command="delete">
                {{ $t("commons.button.delete") }}
              </el-dropdown-item>
              <el-dropdown-item icon="el-icon-delete" :disabled="!onCheckDeletePermissions()" command="delete_force">
                {{ $t("commons.button.delete_force") }}
              </el-dropdown-item>
              <el-dropdown-item icon="el-icon-cpu" command="ai_analysis">
                {{ $t("business.pod.ai_analysis") }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </el-dropdown>
        </template>
      </el-table-column>
    </complex-table>

    <el-dialog
      :title="$t('commons.button.edit')"
      :visible.sync="editDialogVisible"
      width="80%"
      :destroy-on-close="true"
      :close-on-click-modal="false">
      <pod-edit 
        v-if="editDialogVisible"
        :name="selectedPod.name"
        :namespace="selectedPod.namespace"
        @close="closeEditDialog"
        @success="handleEditSuccess">
      </pod-edit>
    </el-dialog>

    <el-dialog
      :title="$t('business.pod.pod_file')"
      :visible.sync="fileDialogVisible"
      width="90%"
      :destroy-on-close="true"
      :close-on-click-modal="false">
      <pod-file-browser
        v-if="fileDialogVisible"
        :name="selectedPod.name"
        :namespace="selectedPod.namespace"
        @close="closeFileDialog">
      </pod-file-browser>
    </el-dialog>

    <el-dialog
      title="Top Pod"
      :visible.sync="topDialogVisible"
      width="80%"
      :destroy-on-close="true"
      :close-on-click-modal="false">
      <pod-top
        v-if="topDialogVisible"
        :key="topDialogKey"
        @close="closeTopDialog"
      >
      </pod-top>
    </el-dialog>

    <el-dialog
      title="终端"
      :visible.sync="terminalDialogVisible"
      width="80%"
      :destroy-on-close="true"
      :close-on-click-modal="false">
      <Terminal
        v-if="terminalDialogVisible"
        :params="terminalDialogParams"
        @close="terminalDialogVisible = false"
      />
    </el-dialog>
    <el-dialog
      title="日志"
      :visible.sync="logDialogVisible"
      width="80%"
      :destroy-on-close="true"
      :close-on-click-modal="false">
      <Terminal
        v-if="logDialogVisible"
        :params="logDialogParams"
        @close="logDialogVisible = false"
      />
    </el-dialog>

    <el-dialog
      :title="$t('business.pod.ai_analysis')"
      :visible.sync="aiAnalysisDialogVisible"
      width="80%"
      :destroy-on-close="true"
      :close-on-click-modal="false">
      <pod-ai-analysis
        v-if="aiAnalysisDialogVisible"
        :name="selectedPod.name"
        :namespace="selectedPod.namespace"
        :cluster="clusterName"
        @close="aiAnalysisDialogVisible = false"
      />
    </el-dialog>

    <el-dialog
      :title="$t('business.event.event')"
      :visible.sync="eventsDialogVisible"
      width="70%"
      :destroy-on-close="true"
      :close-on-click-modal="false">
      <div v-if="selectedPodEvents && selectedPodEvents.length > 0">
        <el-table :data="selectedPodEvents" style="width: 100%" max-height="500">
          <el-table-column prop="type" :label="$t('business.event.type')" width="80">
            <template v-slot:default="{row}">
              <el-tag :type="row.type === 'Warning' ? 'warning' : 'info'" size="mini">
                {{ row.type }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="reason" :label="$t('business.event.reason')" width="120" />
          <el-table-column prop="message" :label="$t('business.event.message')" show-overflow-tooltip />
          <el-table-column prop="firstTimestamp" :label="$t('commons.table.created_time')" width="150">
            <template v-slot:default="{row}">
              {{ row.firstTimestamp | age }}
            </template>
          </el-table-column>
          <el-table-column prop="lastTimestamp" :label="$t('business.event.time')" width="150">
            <template v-slot:default="{row}">
              {{ row.lastTimestamp | age }}
            </template>
          </el-table-column>
          <el-table-column prop="count" :label="$t('business.event.restart')" width="70" />
        </el-table>
      </div>
      <div v-else style="text-align: center; padding: 20px; color: #909399;">
        {{ $t('commons.table.empty_text') }}
      </div>
    </el-dialog>
  </layout-content>
  </div>
</template>

<script>
import LayoutContent from "@/components/layout/LayoutContent"
import {listWorkLoads, deleteWorkLoad, getWorkLoadByName, forceDeleteWorkLoad} from "@/api/workloads"
import {downloadYaml} from "@/utils/actions"
import ComplexTable from "@/components/complex-table"
import {checkPermissions} from "@/utils/permission"
import writeXlsxFile from "write-excel-file";
import { cpuUnitConvert, memoryUnitConvert } from "@/utils/unitConvert"
import { listPodMetrics } from "@/api/apis"
import { searchFullTextItems } from "@/api/fulltextsearch/fulltextsearch"
import { listEventsWithPodSelector } from "@/api/events"
import PodEdit from "./edit"
import PodFileBrowser from "./podfilebrowser"
import PodTop from "./top"
import Terminal from "@/business/workloads/terminal";
import PodAiAnalysis from "./ai-analysis";
export default {
  name: "Pods",
  components: { 
    LayoutContent, 
    ComplexTable,
    PodEdit,
    PodFileBrowser,
    PodTop,
    Terminal,
    PodAiAnalysis
  },
  data () {
    return {
      loading: false,
      data: [],
      paginationConfig: {
        currentPage: 1,
        pageSize: 10,
        total: 0,
      },
      searchConfig: {
        keywords: "",
      },
      selects: [],
      clusterName: "",
      podUsage: [],
      orderField: null,
      orderMethod: null,
      isFullTextSearch: false,
      editDialogVisible: false,
      fileDialogVisible: false,
      topDialogVisible: false,
      topDialogKey: 0, // 新增唯一key
      selectedPod: {
        name: "",
        namespace: ""
      },
      terminalDialogVisible: false,
      logDialogVisible: false,
      terminalDialogParams: {},
      logDialogParams: {},
      aiAnalysisDialogVisible: false,
      eventsDialogVisible: false,
      selectedPodEvents: [],
      podEvents: {}, // 存储Pod事件信息
    }
  },
  methods: {
    openDetail (row) {
      const routeUrl = this.$router.resolve({
        name: "PodDetail",
        params: { namespace: row.metadata.namespace, name: row.metadata.name },
        query: { yamlShow: false, cluster: this.clusterName }
      })
      window.open(routeUrl.href, "_blank")
    },
    onCheckDeletePermissions () {
      return checkPermissions({ scope: "namespace", apiGroup: "", resource: "pods", verb: "delete" })
    },
    checkExecPermissions () {
      return checkPermissions({ scope: 'namespace', apiGroup: '', resource: 'pods/exec', verb: 'create' })
    },
    checkLogPermissions () {
      return checkPermissions({ scope: 'namespace', apiGroup: '', resource: 'pods/log', verb: 'get' })
    },
    handleClick (btn, row) {
      switch (btn) {
        case "download":
          downloadYaml(row.metadata.name + ".yml", getWorkLoadByName(this.clusterName, "pods", row.metadata.namespace, row.metadata.name))
          break
        case "terminal":
          this.openTerminal(row)
          break
        case "logs":
          this.openTerminalLogs(row)
          break
        case "delete":
          this.onDelete(row)
          break
        case "delete_force":
          this.onForceDelete(row)
          break
        case "files":
          this.openPodFiles(row)
          break
        case "ai_analysis":
          this.openAiAnalysis(row)
          break
      }
    },
    onEdit (row) {
      this.selectedPod = {
        name: row.metadata.name,
        namespace: row.metadata.namespace
      }
      this.editDialogVisible = true
    },
    closeEditDialog() {
      this.editDialogVisible = false
    },
    handleEditSuccess() {
      this.editDialogVisible = false
      this.search(true)
      this.$message({
        type: "success",
        message: this.$t("commons.msg.update_success")
      })
    },
    onTop () {
      this.topDialogVisible = false;
      this.$nextTick(() => {
        this.topDialogKey++;
        this.topDialogVisible = true;
      });
    },
    closeTopDialog() {
      this.topDialogVisible = false
    },
    openTerminal (row, container) {
      let c = container || row.containers[0];
      this.terminalDialogParams = {
        cluster: this.clusterName,
        namespace: row.metadata.namespace,
        pod: row.metadata.name,
        container: c,
        type: "terminal"
      };
      this.terminalDialogVisible = true;
    },
    openTerminalLogs (row, container) {
      let c = container || row.containers[0];
      this.logDialogParams = {
        cluster: this.clusterName,
        namespace: row.metadata.namespace,
        pod: row.metadata.name,
        container: c,
        type: "log"
      };
      this.logDialogVisible = true;
    },
    openPodFiles(row, container) {
      let c
      if (container) {
        c = container
      } else {
        c = row.containers[0]
      }
      this.selectedPod = {
        name: row.metadata.name,
        namespace: row.metadata.namespace,
        container: c
      }
      this.fileDialogVisible = true
    },
    closeFileDialog() {
      this.fileDialogVisible = false
    },
    onDelete (row) {
      this.$confirm(this.$t("commons.confirm_message.delete"), this.$t("commons.message_box.prompt"), {
        confirmButtonText: this.$t("commons.button.confirm"),
        cancelButtonText: this.$t("commons.button.cancel"),
        type: "warning",
      }).then(() => {
        this.ps = []
        if (row) {
          this.ps.push(deleteWorkLoad(this.clusterName, "pods", row.metadata.namespace, row.metadata.name))
        } else {
          if (this.selects.length > 0) {
            for (const select of this.selects) {
              this.ps.push(deleteWorkLoad(this.clusterName, "pods", select.metadata.namespace, select.metadata.name))
            }
          }
        }
        if (this.ps.length !== 0) {
          Promise.all(this.ps)
            .then(() => {
              this.search(true)
              this.$message({
                type: "success",
                message: this.$t("commons.msg.delete_success"),
              })
            })
            .catch(() => {
              this.search(true)
            })
        }
      })
    },
    onForceDelete (row) {
      this.$confirm(this.$t("commons.confirm_message.delete"), this.$t("commons.message_box.prompt"), {
        confirmButtonText: this.$t("commons.button.confirm"),
        cancelButtonText: this.$t("commons.button.cancel"),
        type: "warning",
      }).then(() => {
        this.ps = []
        if (row) {
          this.ps.push(forceDeleteWorkLoad(this.clusterName, "pods", row.metadata.namespace, row.metadata.name))
        } else {
          if (this.selects.length > 0) {
            for (const select of this.selects) {
              this.ps.push(forceDeleteWorkLoad(this.clusterName, "pods", select.metadata.namespace, select.metadata.name))
            }
          }
        }
        if (this.ps.length !== 0) {
          Promise.all(this.ps)
            .then(() => {
              this.search(true)
              this.$message({
                type: "success",
                message: this.$t("commons.msg.delete_success"),
              })
            })
            .catch(() => {
              this.search(true)
            })
        }
      })
    },
    onCreate () {
      const routeUrl = this.$router.resolve({ name: "PodCreateYaml", query: { type: "pods", cluster: this.clusterName } })
      window.open(routeUrl.href, "_blank")
    },
    onTop () {
      this.topDialogVisible = true
    },
    getPodStatus (row) {
      if (row.status.containerStatuses) {
        let readyCount = 0
        for (const c of row.status.containerStatuses) {
          if (c.ready) {
            readyCount++
          }
        }
        return readyCount + "/" + row.status.containerStatuses.length
      }
      return
    },
    getRestartTimes (row) {
      if (row.status.containerStatuses) {
        let restartCount = 0
        for (const c of row.status.containerStatuses) {
          restartCount += c.restartCount
        }
        return restartCount
      }
      return 0
    },
    doWithPodList(items){
          let result=[]
          if(items && items.length>0)
          for (const item of items) {
            let container = []
            for (const c of item.spec.containers) {
              container.push(c.name)
            }
            
            item.containers = container
            item.namespace=item.metadata.namespace
            item.status_phase=item.status.phase
            item.pod_status=this.getPodStatus(item)
            item.podIP=item.status.podIP
            item.nodeName=item.spec.nodeName
            item.creationTimestamp=item.metadata.creationTimestamp
            item.restart_count=Number(this.getRestartTimes(item))
            // 添加事件数量字段用于排序
            const eventKey = `${item.metadata.namespace}-${item.metadata.name}`
            item.event_count = this.podEvents[eventKey] ? this.podEvents[eventKey].length : 0
            result.push(item)
          }
          if(!this.orderField || !this.orderMethod){
            return result
          } else{
            let orderMethod=this.orderMethod
             let orderField=this.orderField
            result=result.sort(function(a,b){
              if(orderMethod=='asc'){
                
                return (a[orderField]>b[orderField])?1:-1
              }else{
                return (a[orderField]<b[orderField])?1:-1
              }
            })
            return result
          }
    },
    search (resetPage) {
      this.loading = true
      if (resetPage) {
        this.paginationConfig.currentPage = 1
      }
      if( (!this.orderField || !this.orderMethod ) && !this.isFullTextSearch){
        listWorkLoads(this.clusterName, "pods", true, this.searchConfig.keywords, this.paginationConfig.currentPage, this.paginationConfig.pageSize)
        .then(async (res) => {
          this.data =this.doWithPodList( res.items )
          this.paginationConfig.total = res.total
          // 获取Pod警告事件
          await this.getAllPodsWarningEvents()
        }).finally(() => {
          this.loading = false
        })
      } else {
        let currentPage=this.paginationConfig.currentPage
        let pageSize=this.paginationConfig.pageSize

        listWorkLoads(this.clusterName, "pods", false, "")
        .then(async (res) => {
          let results=[]
          if(!this.isFullTextSearch){
               results = this.doWithPodList( res.items  );
          } else {
               results = this.doWithPodList(searchFullTextItems(res.items,this.searchConfig.keywords));
          } 
          this.data =results.slice(currentPage*pageSize-pageSize,currentPage*pageSize)
          this.paginationConfig.total = results.length
          // 获取Pod警告事件
          await this.getAllPodsWarningEvents()
        }).finally(() => {
          this.loading = false
        })
      }  
    },
    sortTableFun(val){
       this.orderField=null
       this.orderMethod=null
       if (val.order) {
          this.orderField = val.prop
          this.orderMethod = (val.order == "descending" ? "desc" : "asc");
          this.search(false);
       }
    },
    /*导出配额信息为excel*/
    async exportToXlsx(){
      const schema = [
       {
        column: this.$t("commons.table.name"),
        type: String,
        value: (row) => row.metadata.name,
       },
       {
        column: this.$t("business.namespace.namespace"),
        type: String,
        value: (row) => row.metadata.namespace,
       },
       {
        column: "Pod IP",
        type: String,
        value: (row) => row.status.podIP,
       },
       {
        column: "NodeName",
        type: String,
        value: (row) => row.spec.nodeName,
       },
       {
        column: "requests cpu(m)",
        type: Number,
        value: (row) => {
             let result=0
             for(let i=0,s=row.spec.containers.length;i<s;i++){
               result=result+cpuUnitConvert(row.spec.containers[i].resources.requests.cpu)
             }
             return result
        },
       },
       {
        column: "requests memory(Mi)",
        type: Number,
        value: (row) => {
             let result=0
             for(let i=0,s=row.spec.containers.length;i<s;i++){
               result=result+memoryUnitConvert(row.spec.containers[i].resources.requests.memory)
             }
             return result
        },
       },
       {
        column: "limits cpu(m)",
        type: Number,
        value: (row) => {
             let result=0
             for(let i=0,s=row.spec.containers.length;i<s;i++){
               result=result+cpuUnitConvert(row.spec.containers[i].resources.limits.cpu)
             }
             return result
        },
       },
       {
        column: "limits memory(Mi)",
        type: Number,
        value: (row) => {
             let result=0
             for(let i=0,s=row.spec.containers.length;i<s;i++){
               result=result+memoryUnitConvert(row.spec.containers[i].resources.limits.memory)
             }
             return result
        },
       },
       {
        column: "use memory(Mi)",
        type: Number,
        value: (row) => row.memoryUsage,
       },
       {
        column: "use cpu(m)",
        type: Number,
        value: (row) => row.cpuUsage,
       },
      ];
      const data=await listWorkLoads(this.clusterName, "pods", true, this.searchConfig.keywords)
      
      const PodMetrics=await listPodMetrics(this.clusterName)
      const PodMetricsItems=  PodMetrics.items
      const PodMetricsMap={}
          for (const item of PodMetricsItems) {
            let cpu = 0
            let memory = 0
            for (const c of item.containers) {
              cpu += cpuUnitConvert(c.usage.cpu)
              memory += memoryUnitConvert(c.usage.memory)
            }
            PodMetricsMap[item.metadata.namespace+"|"+item.metadata.name] = {
              cpu: cpu,
              memory: memory,
            }
          }
      const pods =data.items;
      for(let i=0,s=pods.length;i<s;i++){
        if(PodMetricsMap[pods[i].metadata.namespace+"|"+pods[i].metadata.name]){
          const m=PodMetricsMap[pods[i].metadata.namespace+"|"+pods[i].metadata.name]
          pods[i].cpuUsage= m.cpu || 0
          pods[i].memoryUsage= m.memory || 0
        }
      }
      await writeXlsxFile((data.items||[]), {
         schema,
         fileName: "pods.xlsx",
      });
    },
    //改变选项"是否全文搜索"
    OnIsFullTextSearchChange(val){
      this.isFullTextSearch=val
    },
    //批量打开终端
    batchTerminal(){
      if (this.selects.length > 0) {
      
              let routeUrl = this.$router.resolve({
                   path: "/batch-terminal",
                   query: {
                    cluster: this.clusterName,
                    terminals: JSON.stringify( this.selects.map((item) => {
                      return {
                         cluster: this.clusterName,
                         namespace: item.metadata.namespace,
                         pod: item.metadata.name,
                         container: item.containers[0],
                         type: "terminal"
                      }
                    }) )
                   }
              })
              window.open(routeUrl.href, "_blank")
      }
    },
    //批量打开日志
    batchLogs(){
      if (this.selects.length > 0) {
      
              let routeUrl = this.$router.resolve({
                   path: "/batch-terminal",
                   query: {
                    cluster: this.clusterName,
                    terminals: JSON.stringify( this.selects.map((item) => {
                      return {
                         cluster: this.clusterName,
                         namespace: item.metadata.namespace,
                         pod: item.metadata.name,
                         container: item.containers[0],
                         type: "log"
                      }
                    }) )
                   }
              })
              window.open(routeUrl.href, "_blank")
      }
    },
    openAiAnalysis(row) {
      this.selectedPod = {
        name: row.metadata.name,
        namespace: row.metadata.namespace
      }
      this.aiAnalysisDialogVisible = true
    },
    // 显示Pod事件详情
    async showPodEvents(row) {
      this.selectedPod = {
        name: row.metadata.name,
        namespace: row.metadata.namespace
      }
      try {
        // 获取该Pod的所有事件（不仅仅是警告）
        const fieldSelector = `involvedObject.name=${row.metadata.name},involvedObject.namespace=${row.metadata.namespace},involvedObject.kind=Pod`
        const events = await listEventsWithPodSelector(this.clusterName, row.metadata.namespace, fieldSelector)
        this.selectedPodEvents = events.items || []
        this.eventsDialogVisible = true
      } catch (error) {
        console.error('获取Pod事件失败:', error)
        this.$message.error('Failed to fetch pod events')
        this.selectedPodEvents = []
        this.eventsDialogVisible = true
      }
    },
    // 获取Pod的警告事件
    async getPodWarningEvents(pod) {
      try {
        const fieldSelector = `involvedObject.name=${pod.metadata.name},involvedObject.namespace=${pod.metadata.namespace},involvedObject.kind=Pod,type=Warning`
        const events = await listEventsWithPodSelector(this.clusterName, pod.metadata.namespace, fieldSelector)
        return events.items || []
      } catch (error) {
        console.error('获取Pod警告事件失败:', error)
        return []
      }
    },
    // 批量获取所有Pod的警告事件
    async getAllPodsWarningEvents() {
      const podEvents = {}
      for (const pod of this.data) {
        const warningEvents = await this.getPodWarningEvents(pod)
        if (warningEvents.length > 0) {
          podEvents[`${pod.metadata.namespace}-${pod.metadata.name}`] = warningEvents
        }
      }
      this.podEvents = podEvents
      // 事件数据加载完成后，重新处理pod列表以更新事件数量字段
      if (this.data && this.data.length > 0) {
        this.data = this.doWithPodList([...this.data])
      }
    },
  },
  mounted () {
    // 确保clusterName是字符串类型，处理cluster参数可能是对象的情况
    const clusterParam = this.$route.query.cluster
    if (typeof clusterParam === 'object' && clusterParam !== null) {
      this.clusterName = clusterParam.name || clusterParam.clusterName || ''
    } else {
      this.clusterName = clusterParam || ''
    }
    this.search()
  },
}
</script>
