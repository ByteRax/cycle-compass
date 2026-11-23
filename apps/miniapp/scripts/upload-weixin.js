/**
 * 微信小程序 CLI 上传脚本
 *
 * 使用方法:
 *   pnpm upload:mp:production                          # 上传 production 构建 (robot: 1)
 *   pnpm upload:mp:development                         # 上传 development 构建 (robot: 2)
 *   pnpm upload:mp:production --version=1.0.1          # 指定版本号（覆盖 package.json）
 *   pnpm upload:mp:development --desc="修复bug"        # 指定版本描述（覆盖 Git commit）
 *   pnpm upload:mp:development --robot=3               # 指定机器人编号（覆盖模式默认值）
 *
 * 版本号策略: 命令行参数 > package.json version
 * 描述策略:   命令行参数 > Git 最新 commit > 默认时间戳
 *
 * 注意事项:
 *   1. 确保已在微信公众平台开通 "小程序代码上传" 权限
 *   2. 确保私钥文件存在（private.${appid}.key），并且配置了上传IP白名单
 *   3. 上传前会自动执行对应 mode 的构建，并跳过打开开发者工具
 *   4. 秘钥文件的appid(VITE_WX_APPID)需要与微信公众平台的小程序appid一致
 */

import { execSync } from 'node:child_process'
import fs from 'node:fs'
import path from 'node:path'
import process from 'node:process'
import { fileURLToPath } from 'node:url'
import ci from 'miniprogram-ci'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const ROOT_DIR = path.resolve(__dirname, '..')
const AUTO_VERSION_BUMP_ON_UPLOAD_MP = 'AUTO_VERSION_BUMP_ON_UPLOAD_MP'
const UPLOAD_MODES = new Set(['development', 'production'])
const DEFAULT_ROBOT_BY_MODE = {
  development: 2,
  production: 1
}

// 从 package.json 读取版本号
function getPackageVersion() {
  try {
    const pkgPath = path.resolve(ROOT_DIR, 'package.json')
    const pkg = JSON.parse(fs.readFileSync(pkgPath, 'utf-8'))
    return pkg.version || '1.0.0'
  } catch {
    return '1.0.0'
  }
}

// 获取最新的 Git commit 信息
function getGitCommitMessage() {
  try {
    const message = execSync('git log -1 --pretty="%an: %s"', {
      cwd: ROOT_DIR,
      encoding: 'utf-8'
    }).trim()
    return message || null
  } catch {
    return null
  }
}

// 生成默认描述
function getDefaultDesc() {
  const gitMessage = getGitCommitMessage()
  if (gitMessage) {
    return gitMessage
  }
  return `上传于 ${new Date().toLocaleString('zh-CN')}`
}

// 解析命令行参数
export function parseArgs(args = process.argv.slice(2)) {
  const params = {
    version: null,
    desc: null,
    mode: null,
    robot: null
  }

  args.forEach((arg) => {
    if (arg.startsWith('--version=')) {
      params.version = arg.split('=')[1]
    } else if (arg.startsWith('--desc=')) {
      params.desc = arg.split('=')[1]
    } else if (arg.startsWith('--mode=')) {
      params.mode = arg.split('=')[1]
    } else if (arg.startsWith('--robot=')) {
      params.robot = Number.parseInt(arg.split('=')[1], 10)
    }
  })

  if (!params.mode || !UPLOAD_MODES.has(params.mode)) {
    throw new Error('请通过 --mode=development 或 --mode=production 指定上传模式')
  }

  if (!params.version) {
    params.version = getPackageVersion()
  }

  if (!params.desc) {
    params.desc = getDefaultDesc()
  }

  params.robot ??= DEFAULT_ROBOT_BY_MODE[params.mode]

  return params
}

export function getBuildCommand(mode) {
  return `pnpm build:mp-weixin:${mode}`
}

// 读取环境变量
function loadEnvFile(mode = 'production') {
  const envPath = path.resolve(ROOT_DIR, 'env', `.env.${mode}`)
  const defaultEnvPath = path.resolve(ROOT_DIR, 'env', '.env')

  const envContent = {}

  // 先读取默认 .env
  if (fs.existsSync(defaultEnvPath)) {
    const content = fs.readFileSync(defaultEnvPath, 'utf-8')
    content.split('\n').forEach((line) => {
      const trimmed = line.trim()
      if (trimmed && !trimmed.startsWith('#')) {
        const [key, ...valueParts] = trimmed.split('=')
        if (key) {
          envContent[key.trim()] = valueParts
            .join('=')
            .trim()
            .replace(/^['"]|['"]$/g, '')
        }
      }
    })
  }

  // 再读取对应模式的 .env 文件（覆盖默认值）
  if (fs.existsSync(envPath)) {
    const content = fs.readFileSync(envPath, 'utf-8')
    content.split('\n').forEach((line) => {
      const trimmed = line.trim()
      if (trimmed && !trimmed.startsWith('#')) {
        const [key, ...valueParts] = trimmed.split('=')
        if (key) {
          envContent[key.trim()] = valueParts
            .join('=')
            .trim()
            .replace(/^['"]|['"]$/g, '')
        }
      }
    })
  }

  return envContent
}

// 获取私钥路径
function getPrivateKeyPath(appid) {
  const searchDirs = [ROOT_DIR, path.resolve(ROOT_DIR, '..')]
  const keyPatterns = [`private.${appid}.key`, 'private.key']

  for (const dir of searchDirs) {
    for (const pattern of keyPatterns) {
      const keyPath = path.resolve(dir, pattern)
      if (fs.existsSync(keyPath)) {
        return keyPath
      }
    }
  }

  throw new Error(`未找到微信小程序上传私钥，请确保存在 private.${appid}.key 文件`)
}

// 主函数
async function main() {
  console.log('\n🚀 开始微信小程序上传流程...\n')

  const params = parseArgs()
  const env = loadEnvFile(params.mode)
  const appid = env.VITE_WX_APPID

  if (!appid) {
    throw new Error('未找到 VITE_WX_APPID 环境变量，请检查 env/.env 文件')
  }

  // 获取私钥路径
  const privateKeyPath = getPrivateKeyPath(appid)
  console.log('🔑 已找到微信小程序上传私钥')

  // 构建小程序（跳过自动打开开发者工具）
  console.log('\n📦 正在构建小程序...（跳过自动打开开发者工具）\n')
  try {
    execSync(getBuildCommand(params.mode), {
      cwd: ROOT_DIR,
      stdio: 'inherit',
      env: {
        ...process.env,
        [AUTO_VERSION_BUMP_ON_UPLOAD_MP]: 'true',
        SKIP_OPEN_DEVTOOLS: 'true'
      }
    })
  } catch (error) {
    console.error('❌ 构建失败:', error.message)
    process.exit(1)
  }
  console.log(`🏷️ 构建模式: ${params.mode}`)
  console.log(`📌 版本号: ${params.version}`)
  console.log(`📝 版本描述: ${params.desc}`)
  console.log(`🤖 机器人编号: ${params.robot}`)

  // 小程序代码目录
  const projectPath = path.resolve(ROOT_DIR, 'dist', 'build', 'mp-weixin')

  if (!fs.existsSync(projectPath)) {
    throw new Error(`构建产物不存在: ${projectPath}`)
  }

  console.log(`📂 项目路径: ${projectPath}`)
  console.log('\n⬆️ 正在上传到微信服务器...\n')

  // 创建项目实例
  const project = new ci.Project({
    appid,
    type: 'miniProgram',
    projectPath,
    privateKeyPath,
    ignores: ['node_modules/**/*']
  })

  try {
    // 上传代码
    const uploadResult = await ci.upload({
      project,
      version: params.version,
      desc: params.desc,
      robot: params.robot,
      setting: {
        es6: true,
        es7: true,
        minify: true,
        autoPrefixWXSS: true,
        minifyWXML: true,
        minifyWXSS: true,
        minifyJS: true
      },
      onProgressUpdate: (task) => {
        if (task._status === 'done') {
          console.log(`  ✅ ${task._msg}`)
        }
      }
    })

    console.log('\n✅ 上传成功!')
    console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━')
    console.log(`  📌 版本号: ${params.version}`)
    console.log(`  📝 描述: ${params.desc}`)
    console.log(`  🤖 机器人: ${params.robot}`)
    console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━')
    console.log('\n📋 下一步操作:')
    console.log('  1. 登录微信公众平台: https://mp.weixin.qq.com')
    console.log('  2. 进入 "管理 -> 版本管理"')
    console.log('  3. 在 "开发版本" 中找到刚上传的版本')
    console.log('  4. 点击 "选为体验版" 按钮\n')

    return uploadResult
  } catch (error) {
    console.error('\n❌ 上传失败:', error.message)
    if (error.message.includes('privateKeyPath')) {
      console.log('\n💡 提示: 请确保已在微信公众平台配置代码上传密钥')
      console.log('   1. 登录微信公众平台')
      console.log('   2. 进入 "开发 -> 开发设置"')
      console.log('   3. 在 "小程序代码上传" 区域生成并下载密钥')
      console.log('   4. 在 "小程序代码上传" 区域配置上传IP白名单')
    }
    process.exit(1)
  }
}

if (path.resolve(process.argv[1] || '') === __filename) {
  main().catch((error) => {
    console.error('❌ 执行出错:', error)
    process.exit(1)
  })
}
