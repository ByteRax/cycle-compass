export interface UploadParams {
  desc: string
  mode: 'development' | 'production'
  robot: number
  version: string
}

export function getBuildCommand(mode: UploadParams['mode']): string
export function parseArgs(args?: string[]): UploadParams
