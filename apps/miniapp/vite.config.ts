import type { ConfigEnv } from 'vite'
import * as path from 'node:path'
import * as process from 'node:process'
import CompressJson from '@binbinji/unplugin-compress-json/vite'
import Uni from '@uni-helper/plugin-uni'
import UniComponents from '@uni-helper/vite-plugin-uni-components'
import UniHelperManifest from '@uni-helper/vite-plugin-uni-manifest'
import UniPages from '@uni-helper/vite-plugin-uni-pages'
import UniOptimization from '@uni-ku/bundle-optimizer'
import UniKuRoot from '@uni-ku/root'
import { codeInspectorPlugin } from 'code-inspector-plugin'
import { visualizer } from 'rollup-plugin-visualizer'
import AutoImport from 'unplugin-auto-import/vite'
import { defineConfig, loadEnv } from 'vite'
import viteImagemin from 'vite-plugin-imagemin'
import ViteRestart from 'vite-plugin-restart'
import { WeappTailwindcss } from 'weapp-tailwindcss/vite'
import { handlePageName, handleSubPackagePageName } from './vite-plugins/vite-config-uni-pages'
import openDevTools from './vite-plugins/vite-open-dev-tools'
import { AutoVersion } from './vite-plugins/vite-plugin-auto-version'
import { isBuildCommand } from './getEnv'
import manifestConfig from './manifest.config'

// https://vitejs.dev/config/
export default async ({ mode }: ConfigEnv) => {
  const { UNI_PLATFORM, SKIP_OPEN_DEVTOOLS } = process.env
  const rawEnv = loadEnv(mode, path.resolve(process.cwd(), 'env'))
  const env: Record<string, any> = {
    VITE_APP_TITLE: '周期罗盘',
    VITE_APP_PUBLIC_BASE: '/',
    ...rawEnv
  }
  process.env['VITE_APP_TITLE'] = env['VITE_APP_TITLE']
  const {
    VITE_APP_PORT,
    VITE_SERVER_BASEURL,
    VITE_APP_TITLE,
    VITE_APP_PUBLIC_BASE,
    VITE_APP_PROXY_ENABLE,
    VITE_APP_PROXY_PREFIX
  } = env
  const isBuild = isBuildCommand()
  console.warn('UNI_PLATFORM -> ', UNI_PLATFORM)

  return defineConfig({
    envDir: './env',
    base: VITE_APP_PUBLIC_BASE,

    plugins: [
      AutoVersion({
        type: 'patch',
        inject: true
      }),
      await UniHelperManifest(),
      UniPages({
        dts: 'src/types/uni-pages.d.ts',
        subPackages: ['src/pages-sub'],
        onAfterMergePageMetaData: (pageMetaData, subPageMetaData) => {
          handlePageName(pageMetaData)
          handleSubPackagePageName(subPageMetaData)
        },
        exclude: ['**/components/**/*.*', '**/layout/**/*.*']
      }),
      UniComponents({
        extensions: ['vue'],
        deep: true,
        directoryAsNamespace: false,
        dts: 'src/types/components.d.ts',
        dirs: ['src/components', 'src/layout']
      }),
      UniKuRoot({
        enabledGlobalRef: true
      }),
      Uni(),
      WeappTailwindcss({
        cssEntries: [path.resolve(__dirname, 'src/app.css')],
        rem2rpx: true
      }),
      {
        name: 'fix-vite-plugin-vue',
        configResolved(config) {
          const plugin = config.plugins.find((p) => p.name === 'vite:vue')
          if (plugin && plugin.api && plugin.api.options) {
            plugin.api.options.devToolsEnabled = false
          }
        }
      },
      AutoImport({
        imports: ['vue', 'pinia', 'uni-app'],
        dts: 'src/types/auto-import.d.ts',
        dirs: ['src/composables/**', 'src/stores/**', 'src/utils/**', 'src/hooks/**', 'src/router/**'],
        eslintrc: {
          enabled: true,
          globalsPropValue: true
        },
        vueTemplate: true
      }),
      UniOptimization({
        enable: {
          'optimization': true,
          'async-import': true,
          'async-component': true
        },
        logger: false
      }),
      isBuild &&
        viteImagemin({
          gifsicle: {
            optimizationLevel: 7,
            interlaced: false
          },
          optipng: {
            optimizationLevel: 7
          },
          webp: {
            quality: 85
          },
          mozjpeg: {
            quality: 85
          },
          pngquant: {
            quality: [0.8, 0.9],
            speed: 4
          },
          svgo: {
            plugins: [{ name: 'removeViewBox' }, { name: 'removeEmptyAttrs', active: false }]
          }
        }),
      CompressJson(),
      ViteRestart({
        restart: ['vite.config.ts']
      }),
      !isBuild &&
        codeInspectorPlugin({
          showSwitch: true,
          autoToggle: false,
          bundler: 'vite'
        }),
      UNI_PLATFORM === 'h5' && {
        name: 'html-transform',
        transformIndexHtml: {
          order: 'post',
          handler(html) {
            const title = manifestConfig.h5?.title || manifestConfig.name || VITE_APP_TITLE
            const buildTime = new Date().toISOString().replace('T', ' ').slice(0, 19)
            return html
              .replace('%BUILD_TIME%', buildTime)
              .replace(/<title>[\s\S]*?<\/title>/, () => `<title>${title}</title>`)
          }
        }
      },
      UNI_PLATFORM === 'h5' &&
        isBuild &&
        visualizer({
          filename: './node_modules/.cache/visualizer/stats.html',
          open: process.env['VISUALIZER_OPEN'] === 'true',
          gzipSize: true,
          brotliSize: true
        }),
      SKIP_OPEN_DEVTOOLS !== 'true' && openDevTools({ mode })
    ],
    define: {
      __UNI_PLATFORM__: JSON.stringify(UNI_PLATFORM),
      __VITE_APP_PROXY__: JSON.stringify(VITE_APP_PROXY_ENABLE)
    },
    assetsInclude: ['**/*.md'],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
        '@components': path.resolve(__dirname, './src/components'),
        '@layout': path.resolve(__dirname, './src/layout'),
        '@utils': path.resolve(__dirname, './src/utils')
      },
      extensions: ['.js', '.ts', '.jsx', '.tsx']
    },
    server: {
      host: '0.0.0.0',
      open: !isBuild,
      hmr: true,
      port: Number.parseInt(`${VITE_APP_PORT}`, 10),
      proxy:
        VITE_APP_PROXY_ENABLE === 'true'
          ? {
              [`${VITE_APP_PROXY_PREFIX}`]: {
                target: VITE_SERVER_BASEURL,
                changeOrigin: true,
                rewrite: (path) => path.replace(new RegExp(`^${VITE_APP_PROXY_PREFIX}`), '')
              }
            }
          : undefined,
      warmup: {
        clientFiles: ['./index.html', './src/{views,components,layout}/**/*.{vue,ts,js}']
      }
    },
    esbuild: {
      drop: ['debugger']
    },
    build: {
      sourcemap: !isBuild,
      target: 'es2020',
      cssTarget: 'chrome61',
      cssCodeSplit: true,
      cssMinify: isBuild ? 'esbuild' : false,
      minify: isBuild ? 'terser' : false,
      terserOptions: {
        compress: isBuild
          ? {
              ecma: 2020,
              drop_console: true,
              drop_debugger: true,
              pure_funcs: ['console.log', 'console.debug', 'console.warn'],
              pure_getters: true,
              keep_fargs: false,
              unsafe_comps: false,
              unsafe: false,
              unsafe_math: false
            }
          : false
      },
      rollupOptions: {
        output: isBuild
          ? {
              experimentalMinChunkSize: 10000,
              chunkFileNames: 'static/js/[name]-[hash].js',
              entryFileNames: 'static/js/[name]-[hash].js',
              assetFileNames: 'static/[ext]/[name]-[hash][extname]',
              compact: true
            }
          : {},
        treeshake: {
          moduleSideEffects: true
        }
      }
    }
  })
}
