import { describe, expect, it } from 'vitest'

import { getBuildCommand, parseArgs } from '../../scripts/upload-weixin.js'

describe('微信小程序上传脚本配置与参数解析', () => {
  it.each([
    ['production', 1],
    ['development', 2]
  ])('%s 模式使用对应的默认 robot 和构建命令', (mode, robot) => {
    const params = parseArgs([`--mode=${mode}`, '--desc=测试上传'])

    expect(params.robot).toBe(robot)
    expect(getBuildCommand(params.mode)).toBe(`pnpm build:mp-weixin:${mode}`)
  })

  it('允许显式覆盖默认 robot', () => {
    expect(parseArgs(['--mode=development', '--desc=测试上传', '--robot=3']).robot).toBe(3)
  })

  it('允许显式指定版本号与描述', () => {
    const params = parseArgs(['--mode=production', '--version=2.0.0', '--desc=功能上线'])
    expect(params.version).toBe('2.0.0')
    expect(params.desc).toBe('功能上线')
  })

  it('拒绝未指定或无效的上传模式', () => {
    expect(() => parseArgs(['--desc=测试上传'])).toThrow('请通过 --mode=development 或 --mode=production 指定上传模式')
    expect(() => parseArgs(['--mode=test', '--desc=测试上传'])).toThrow('请通过 --mode=development 或 --mode=production 指定上传模式')
  })
})
