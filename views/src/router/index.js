import Vue from 'vue'
import Router from 'vue-router'
import Index from '@/components/Index'
import Redis from '@/components/CacheIndex'
Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      name: 'Index',
      component: Index
    },
    {
      path : '/redis',
      name : 'Redis',
      component : Redis
    }
  ]
})
