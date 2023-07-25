import Vue from 'vue'
import Router from 'vue-router'
import Index from '@/components/Index'
import Home from '@/components/Home'
Vue.use(Router)

export default new Router({
  routes: [
    {
      path : '/',
      name : 'home',
      mode : 'hash', //使用hash模式 本地路由
      component : Home,
      children : [
        {
          path : '/CacheIndex',
          components : {
            home : () => import('../components/CacheIndex') //指定其中的home  router view
          },
          meta: { keepAlive: true }
        },
        {
          path : '/Consumer',
          components : {
            home : () => import('../components/Consumer') //指定其中的home  router view
          },
          meta: { keepAlive: true }
        },
        {
          path : '/Git',
          components : {
            home : () => import('../components/Git') //指定其中的home  router view
          },
          meta: { keepAlive: true }
        },
        {
          path : '/WechatKefu',
          components : {
            home : () => import('../components/WechatKefu') //指定其中的home  router view
          },
          meta: { keepAlive: true }
        },
        {
          path : '/Vip',
          components : {
            home : () => import('../components/Vip') //指定其中的home  router view
          },
          meta: { keepAlive: true }
        },
        {
          path : '/Link',
          components : {
            home : () => import('../components/Link') //指定其中的home  router view
          },
          meta: { keepAlive: true }
        },
        {
          path : '/Docker',
          components : {
            home : () => import('../components/Docker') //指定其中的home  router view
          },
          meta: { keepAlive: true }
        },
        {
          path : '/Model',
          components : {
            home : () => import('../components/Model') //指定其中的home  router view
          },
          meta: { keepAlive: true }
        },
        {
          path : '/Ssh',
          components : {
            home : () => import('../components/Ssh') //指定其中的home  router view
          },
          meta: { keepAlive: true }
        }
      ]

    }
  ]
})
