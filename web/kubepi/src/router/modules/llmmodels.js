import Layout from "@/business/app-layout/horizontal-layout";

const LLMModels = {
  path: "/llmmodels",
  sort: 5,
  component: Layout,
  name: "LLMModels",
  requirePermission: {
    resource: "llmmodels",
    verb: "list"
  },
  meta: {
    title: "business.llmmodels.llm_model",
    icon: "el-icon-s-platform",
  },
  children: [
    {
      path: "/llmmodels",
      component: () => import("@/business/llmmodels/index"),
      name: "LLMModelList",
      meta: {
        title: "business.llmmodels.llm_model_list",
        activeMenu: "/llmmodels",
      },
    },
    {
      path: "/llmmodels/create",
      component: () => import("@/business/llmmodels/Create"),
      name: "CreateLLMModel",
      meta: {
        title: "business.llmmodels.create_llm_model",
        activeMenu: "/llmmodels",
      },
      hidden: true,
    },
    {
      path: "/llmmodels/edit/:name",
      component: () => import("@/business/llmmodels/Edit"),
      name: "EditLLMModel",
      meta: {
        title: "business.llmmodels.edit_llm_model",
        activeMenu: "/llmmodels",
      },
      hidden: true,
    },
    {
      path: "/llmmodels/test/:name",
      component: () => import("@/business/llmmodels/Test"),
      name: "TestLLMModel",
      meta: {
        title: "business.llmmodels.test_llm_model",
        activeMenu: "/llmmodels",
      },
      hidden: true,
    }
  ],
}

export default LLMModels; 