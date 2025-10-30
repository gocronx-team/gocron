<template>
   <el-container>
    <el-header v-if="userStore.isLogin">
      <app-header></app-header>
      <app-nav-menu></app-nav-menu>
    </el-header>
    <el-main>
      <div id="main-container" v-cloak>
        <router-view v-slot="{ Component }">
          <keep-alive>
            <component :is="Component" />
          </keep-alive>
        </router-view>
      </div>
    </el-main>
    <el-footer v-if="!isFooterEmpty">
      <app-footer ref="footerRef"></app-footer>
    </el-footer>
  </el-container>
</template>

<script setup>
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from './stores/user'
import installService from './api/install'
import appHeader from './components/common/header.vue'
import appNavMenu from './components/common/navMenu.vue'
import appFooter from './components/common/footer.vue'

const router = useRouter()
const userStore = useUserStore()
const footerRef = ref(null)

const isFooterEmpty = computed(() => {
  return footerRef.value?.isEmpty ?? true
})

onMounted(() => {
  installService.status((data) => {
    if (!data) {
      router.push('/install')
    }
  })
})
</script>

<style>
[v-cloak] {
  display: none !important;
}
html, body {
  margin: 0;
  padding: 0;
  height: 100%;
  overflow-x: hidden;
}
.el-header {
  padding: 0;
  margin: 0;
}
.el-container {
  padding: 0;
  margin: 0;
  width: 100%;
}
.el-main {
  padding: 0;
  margin: 0;
}
#main-container .el-main {
  height: calc(100vh - 116px);
  margin: 20px 20px 0 20px;
}
.el-header + .el-main #main-container .el-main {
  height: calc(100vh - 116px);
}
.el-main:first-child #main-container .el-main {
  height: 100vh;
  margin: 0;
}

.el-aside .el-menu {
  height: 100%;
}
.custom-message-box {
  min-width: 420px;
}
.custom-message-box .el-message-box__message {
  font-size: 15px;
  line-height: 1.6;
}
.el-message-box__title {
  font-size: 18px;
  font-weight: 600;
}
</style>
