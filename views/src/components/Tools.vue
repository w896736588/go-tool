<template>
  <el-card>
    <div style="margin-bottom: 20px;">
      <el-button
        size="small"
        @click="addTab(editableTabsValue)"
      >
        新页卡
      </el-button>
    </div>
<!--    如果想关闭 那么增加一个属性 closable-->
    <el-tabs v-model="editableTabsValue" type="card"  @tab-remove="removeTab">
      <el-tab-pane
        v-for="(item, index) in editableTabs"
        :key="item.name"
        :label="item.title"
        :name="item.name"
      >

        <el-card>
          <el-radio @change="editableTabs[index].title = 'json格式化';" v-model="toolType" label="1">json格式化</el-radio>
          <el-radio v-model="toolType" label="2">时间戳转换</el-radio>
          <el-radio v-model="toolType" label="3">二维码生成</el-radio>
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </el-card>
</template>

<script>
export default {
  name: "tools",
  data() {
    return {
      editableTabsValue: '1',
      editableTabs: [{
        title: 'Tab 1',
        name: '1',
        content: 'Tab 1 content'
      }],
      tabIndex: 1,

      //业务
      toolType : '',
    }
  },
  methods : {
    addTab(targetName) {
      let newTabName = ++this.tabIndex + '';
      this.editableTabs.push({
        title: 'new',
        name: newTabName,
        content: 'New Tab content'
      });
      this.editableTabsValue = newTabName;
    },
    removeTab(targetName) {
      let tabs = this.editableTabs;
      let activeName = this.editableTabsValue;
      if (activeName === targetName) {
        tabs.forEach((tab, index) => {
          if (tab.name === targetName) {
            let nextTab = tabs[index + 1] || tabs[index - 1];
            if (nextTab) {
              activeName = nextTab.name;
            }
          }
        });
      }

      this.editableTabsValue = activeName;
      this.editableTabs = tabs.filter(tab => tab.name !== targetName);
    }
  },
}
</script>

<style scoped>

</style>
