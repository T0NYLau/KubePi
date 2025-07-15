<template>
  <div class="app-container">
    <el-card class="box-card" shadow="never">
      <div slot="header" class="clearfix">
        <span>{{ $t('business.llmmodels.llm_model_list') }}</span>
        <el-button style="float: right; padding: 3px 0" type="text" @click="handleCreate">
          <i class="el-icon-plus"></i> {{ $t('commons.button.create') }}
        </el-button>
      </div>
      <div class="table-search">
        <el-input
          v-model="queryCondition.name"
          :placeholder="$t('commons.search.quickSearch')"
          prefix-icon="el-icon-search"
          @keyup.enter.native="search"
          clearable
          @clear="clearSearch"
          style="width: 200px;"
        />
        <el-button type="primary" size="mini" @click="search">{{ $t('commons.button.search') }}</el-button>
      </div>

      <el-table
        v-loading="loading"
        :data="tableData"
        border
        style="width: 100%"
        @sort-change="sortChange"
      >
        <el-table-column
          prop="name"
          :label="$t('business.llmmodels.name')"
        />
        <el-table-column
          prop="baseUri"
          :label="$t('business.llmmodels.base_uri')"
        />
        <el-table-column
          prop="modelName"
          :label="$t('business.llmmodels.model_name')"
        />
        <el-table-column
          prop="temperature"
          :label="$t('business.llmmodels.temperature')"
        />
        <el-table-column
          prop="status"
          :label="$t('business.llmmodels.status')"
        >
          <template slot-scope="scope">
            <el-tag :type="getStatusType(scope.row.status)">
              {{ $t('business.llmmodels.' + scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('business.llmmodels.created_at')"
        >
          <template slot-scope="scope">
            {{ scope.row.createdAt | moment('YYYY-MM-DD HH:mm:ss') }}
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('commons.table.action')"
          width="250"
        >
          <template slot-scope="scope">
            <el-button
              size="mini"
              type="primary"
              @click="handleTest(scope.row)"
            >{{ $t('business.llmmodels.test_connection') }}</el-button>
            <el-button
              size="mini"
              type="success"
              @click="handleEdit(scope.row)"
            >{{ $t('commons.button.edit') }}</el-button>
            <el-button
              size="mini"
              type="danger"
              @click="handleDelete(scope.row)"
            >{{ $t('commons.button.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          background
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
          :current-page="currentPage"
          :page-sizes="[10, 20, 30, 50]"
          :page-size="pageSize"
          layout="total, sizes, prev, pager, next, jumper"
          :total="total">
        </el-pagination>
      </div>
    </el-card>
  </div>
</template>

<script>
import { searchLLMModels, deleteLLMModel } from '@/api/llmmodels';

export default {
  name: 'LLMModelsList',
  data() {
    return {
      loading: false,
      tableData: [],
      queryCondition: {
        name: ''
      },
      currentPage: 1,
      pageSize: 10,
      total: 0,
    };
  },
  created() {
    this.search();
  },
  methods: {
    getStatusType(status) {
      switch(status) {
        case 'available':
          return 'success';
        case 'unavailable':
          return 'danger';
        case 'untested':
          return 'info';
        default:
          return '';
      }
    },
    search() {
      this.loading = true;
      const conditions = [];
      if (this.queryCondition.name) {
        conditions.push({
          field: 'name',
          operator: 'like',
          value: this.queryCondition.name
        });
      }
      searchLLMModels(this.currentPage, this.pageSize, conditions)
        .then(response => {
          this.tableData = response.data.items;
          this.total = response.data.total;
        })
        .catch(error => {
          console.error(error);
          this.$message.error(this.$t('commons.msg.search_failed'));
        })
        .finally(() => {
          this.loading = false;
        });
    },
    clearSearch() {
      this.queryCondition.name = '';
      this.search();
    },
    sortChange(column) {
      // Handle sorting if needed
    },
    handleSizeChange(size) {
      this.pageSize = size;
      this.search();
    },
    handleCurrentChange(current) {
      this.currentPage = current;
      this.search();
    },
    handleCreate() {
      this.$router.push('/llmmodels/create');
    },
    handleEdit(row) {
      this.$router.push(`/llmmodels/edit/${row.name}`);
    },
    handleTest(row) {
      this.$router.push(`/llmmodels/test/${row.name}`);
    },
    handleDelete(row) {
      this.$confirm(
        this.$t('commons.confirm_message.delete'),
        this.$t('commons.message_box.confirm'),
        {
          confirmButtonText: this.$t('commons.button.confirm'),
          cancelButtonText: this.$t('commons.button.cancel'),
          type: 'warning'
        }
      ).then(() => {
        deleteLLMModel(row.name)
          .then(() => {
            this.$message.success(this.$t('commons.msg.delete_success'));
            this.search();
          })
          .catch(error => {
            console.error(error);
            this.$message.error(this.$t('commons.msg.delete_failed'));
          });
      }).catch(() => {
        // Cancel deletion
      });
    }
  }
};
</script>

<style scoped>
.table-search {
  margin-bottom: 20px;
}
.pagination {
  margin-top: 20px;
  text-align: right;
}
</style> 