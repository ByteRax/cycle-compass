import type {
  BacktestOutput,
  CombinedSignal,
  CycleOverview,
  IndustryMetric,
  MarketQuote,
  PortfolioOverview
} from '@/api/cycle'

export interface PriceItem {
  name: string
  price: number
  change_pct: number
}

export interface SymbolPriceItem extends PriceItem {
  symbol: string
}

export interface FxQuoteItem {
  symbol: string
  name: string
  price: string
  change_pct: number
  currency: string
}

export interface UserPortfolioTx {
  id?: string
  date: string
  week_key?: string
  event_type: string
  target_name: string
  symbol?: string | null
  amount: number
  description: string
  shares?: number
  price?: number
  cash_delta?: number
  cash_pool_delta?: number
  cash_balance_after?: number
  cash_pool_after?: number
  source?: string
  unverified?: boolean
  [key: string]: any
}

export interface UserPortfolioState {
  account_id: string
  weekly_amount: number
  available_cash: number
  cash_pool: number
  dividend_shares: number
  dividend_cost: number
  growth_shares: number
  growth_cost: number
  transactions: UserPortfolioTx[]
}

export const SEED_GLOBAL_ECONOMIC_DATA: MarketQuote[] = [
  { symbol: 'BRENT', name: '布伦特原油', category: 'global', price: 87.19, change: 0.50, change_pct: 0.58, currency: 'USD', as_of: '2026-11-24', meta: { source_name: 'ICE/Yahoo', symbol: 'BRENT', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { symbol: 'VIX', name: '恐慌指数', category: 'global', price: 28.67, change: 2.06, change_pct: 7.74, currency: '', as_of: '2026-11-24', meta: { source_name: 'CBOE/Yahoo', symbol: 'VIX', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { symbol: 'DXY', name: '美元指数', category: 'fx', price: 103.98, change: 0.04, change_pct: 0.04, currency: 'USD', as_of: '2026-11-24', meta: { source_name: 'ICE/Yahoo', symbol: 'DXY', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { symbol: 'US20Y', name: '美债长债', category: 'bond', price: 81.69, change: 0.24, change_pct: 0.29, currency: 'USD', as_of: '2026-11-24', meta: { source_name: 'CBOE/Yahoo', symbol: 'US20Y', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { symbol: 'GOLD_OZ', name: '黄金盎司', category: 'commodity', price: 3624.33, change: -10.87, change_pct: -0.30, currency: 'USD', as_of: '2026-11-24', meta: { source_name: 'COMEX/Yahoo', symbol: 'GOLD_OZ', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { symbol: 'ICNYXAU', name: '人民币账户黄金', category: 'commodity', price: 628.50, change: 2.15, change_pct: 0.34, currency: '元/克', as_of: '2026-11-24', meta: { source_name: '工行/新浪', symbol: 'ICNYXAU', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { symbol: 'SILVER_OZ', name: '白银盎司', category: 'commodity', price: 42.24, change: -0.91, change_pct: -2.11, currency: 'USD', as_of: '2026-11-24', meta: { source_name: 'COMEX/Yahoo', symbol: 'SILVER_OZ', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { symbol: 'COPPER_HG', name: '铜', category: 'commodity', price: 551.57, change: -20.65, change_pct: -3.61, currency: 'USD', as_of: '2026-11-24', meta: { source_name: 'COMEX/Yahoo', symbol: 'COPPER_HG', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { symbol: 'NATGAS', name: '天然气', category: 'commodity', price: 3.60, change: 0.08, change_pct: 2.21, currency: 'USD', as_of: '2026-11-24', meta: { source_name: 'NYMEX/Yahoo', symbol: 'NATGAS', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } }
]

export const SEED_SECTOR_LIST: IndustryMetric[] = [
  { sector_id: 'ai_compute', name: 'AI算力', price_change_pct: -4.22, stage: '繁荣期', prosperity_score: 95.0, price_index: 95.78, unit: '点', catalyst: '前沿大模型持续迭代，美股盘前震荡', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'ai_compute', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'cpo', name: 'CPO', price_change_pct: -3.57, stage: '繁荣期', prosperity_score: 92.5, price_index: 96.43, unit: '点', catalyst: '1.6T光模块与硅光渗透加速交付', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'cpo', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'semiconductor', name: '半导体', price_change_pct: -3.93, stage: '复苏上行', prosperity_score: 88.0, price_index: 96.07, unit: '点', catalyst: '先进制程与汽车芯片需求拉动', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'semiconductor', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'storage', name: '存储', price_change_pct: -5.39, stage: '景气高位', prosperity_score: 90.0, price_index: 94.61, unit: '点', catalyst: 'HBM与大容量DRAM/NAND价格坚挺', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'storage', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'datacenter', name: '数据中心', price_change_pct: -1.76, stage: '稳步扩张', prosperity_score: 84.0, price_index: 98.24, unit: '点', catalyst: '智算中心IDC规模化交付上架', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'datacenter', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'cloud', name: '云计算', price_change_pct: -1.70, stage: '成熟增长', prosperity_score: 80.0, price_index: 98.30, unit: '点', catalyst: '企业AI工作负载云化转型', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'cloud', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'commercial_aerospace', name: '商业航天', price_change_pct: -2.12, stage: '早期爆发', prosperity_score: 78.0, price_index: 97.88, unit: '点', catalyst: '商业火箭发射密集，商业化在即', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'commercial_aerospace', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'satellite', name: '卫星', price_change_pct: -0.80, stage: '早期爆发', prosperity_score: 77.0, price_index: 99.20, unit: '点', catalyst: '低轨互联网星座批量组网', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'satellite', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'robotics', name: '机器人', price_change_pct: -1.89, stage: '技术突破', prosperity_score: 82.0, price_index: 98.11, unit: '点', catalyst: '具身智能本体与轻量关节迭代', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'robotics', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'autonomous_driving', name: '自动驾驶', price_change_pct: -1.63, stage: '规模试点', prosperity_score: 85.0, price_index: 98.37, unit: '点', catalyst: '端到端智驾与Robotaxi商业化', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'autonomous_driving', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'nuclear_power', name: '核电', price_change_pct: -1.73, stage: '稳健扩张', prosperity_score: 76.0, price_index: 98.27, unit: '点', catalyst: '清洁基荷能源与三代四代核电开工', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'nuclear_power', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'power_grid', name: '电网', price_change_pct: -0.95, stage: '建设高峰', prosperity_score: 79.0, price_index: 99.05, unit: '点', catalyst: '特高压外送通道与配电网智能化改造', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'power_grid', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'military', name: '军工', price_change_pct: 0.06, stage: '周期底部', prosperity_score: 72.0, price_index: 100.06, unit: '点', catalyst: '新型号列装与订单复苏预期', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'military', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'new_energy', name: '新能源', price_change_pct: -1.40, stage: '出清修复', prosperity_score: 68.0, price_index: 98.60, unit: '点', catalyst: '风光储一体化海外需求稳健', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'new_energy', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'solar', name: '光伏', price_change_pct: 0.25, stage: '供给见底', prosperity_score: 71.0, price_index: 100.25, unit: '点', catalyst: '产业链价格企稳与落后产能加速出清', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'solar', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'lithium_battery', name: '锂电池', price_change_pct: -0.85, stage: '逐步筑底', prosperity_score: 73.0, price_index: 99.15, unit: '点', catalyst: '储能出货爆发与碳酸锂库存去化', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'lithium_battery', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'petroleum', name: '石油', price_change_pct: 0.70, stage: '高位震荡', prosperity_score: 75.0, price_index: 100.70, unit: '点', catalyst: 'OPEC+受控减产与地缘溢价支持', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'petroleum', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'natural_gas', name: '天然气', price_change_pct: 2.21, stage: '震荡下行', prosperity_score: 65.0, price_index: 102.21, unit: '点', catalyst: '库存相对充裕与季节性淡季', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'natural_gas', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'copper_metal', name: '铜 / 有色', price_change_pct: 0.02, stage: '高位景气', prosperity_score: 83.0, price_index: 100.02, unit: '点', catalyst: '电网改造与AI能耗高用铜驱动', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'copper_metal', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'gold', name: '黄金', price_change_pct: 1.05, stage: '避险繁荣', prosperity_score: 89.0, price_index: 101.05, unit: '点', catalyst: '央行持续购金与全球去美元化叙事', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'gold', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'banking', name: '银行金融', price_change_pct: 0.11, stage: '防守稳定', prosperity_score: 69.0, price_index: 100.11, unit: '点', catalyst: '高股息资产防御与资产质量稳健', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'banking', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'biomedicine', name: '生物医药', price_change_pct: -0.09, stage: '估值消化', prosperity_score: 67.0, price_index: 99.91, unit: '点', catalyst: '创新药出海提速但短期震荡', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'biomedicine', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'consumption', name: '消费', price_change_pct: 0.26, stage: '弱复苏', prosperity_score: 64.0, price_index: 100.26, unit: '点', catalyst: '以旧换新政策驱动与性价比消费', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'consumption', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } },
  { sector_id: 'rare_earth', name: '稀土', price_change_pct: -0.15, stage: '底部企稳', prosperity_score: 74.0, price_index: 99.85, unit: '点', catalyst: '出口配额调控与人形机器人磁材需求', as_of: '2026-11-24', meta: { source_name: '离线参考基准', symbol: 'rare_earth', as_of: '2026-11-24', status: 'offline_seed', updated_at: '' } }
]

export const SEED_KR_MACRO = [
  { symbol: 'KOSPI', name: 'KOSPI', price: 2510.50, change_pct: 6.32 },
  { symbol: 'KOSDAQ', name: 'KOSDAQ', price: 720.10, change_pct: 4.79 }
]

export const SEED_KR_SECTORS = [
  { name: '存储', price: 106.70, change_pct: 6.70 },
  { name: '半导体', price: 106.73, change_pct: 6.73 },
  { name: '电池', price: 107.89, change_pct: 7.89 },
  { name: '消费电子', price: 116.10, change_pct: 16.10 },
  { name: '互联网', price: 104.33, change_pct: 4.33 },
  { name: '汽车', price: 111.52, change_pct: 11.52 },
  { name: '生物医药', price: 103.71, change_pct: 3.71 },
  { name: '化工材料', price: 111.11, change_pct: 11.11 }
]

export const SEED_JP_MACRO = [
  { symbol: 'N225', name: '日经225', price: 38283.85, change_pct: 3.91 },
  { symbol: 'TOPIX', name: 'TOPIX', price: 2680.30, change_pct: 2.47 }
]

export const SEED_JP_SECTORS = [
  { name: '半导体设备', price: 107.28, change_pct: 7.28 },
  { name: '工业自动化', price: 105.88, change_pct: 5.88 },
  { name: '精密制造', price: 105.17, change_pct: 5.17 },
  { name: '汽车产业链', price: 99.97, change_pct: -0.03 },
  { name: '消费电子', price: 100.23, change_pct: 0.23 },
  { name: '半导体材料', price: 104.24, change_pct: 4.24 },
  { name: '电子元件', price: 114.89, change_pct: 14.89 },
  { name: '游戏娱乐', price: 98.01, change_pct: -1.99 }
]

export const SEED_ASIA_MACRO = [
  { symbol: 'VN_INDEX', name: '越南胡志明', price: 1245.20, change_pct: -0.05 },
  { symbol: 'SENSEX', name: '孟买SENSEX', price: 79540.0, change_pct: 2.30 }
]

export const SEED_HK_US_MACRO = [
  { symbol: 'HSI', name: '恒生指数', price: 22219.65, change_pct: -0.73 },
  { symbol: 'HSTECH', name: '恒生科技', price: 3466.87, change_pct: -1.17 },
  { symbol: 'HSCE', name: '国企指数', price: 7452.56, change_pct: -1.00 },
  { symbol: 'FXI', name: '富时中国50', price: 31.63, change_pct: -0.63 },
  { symbol: 'KWEB', name: '中概互联', price: 21.41, change_pct: -0.79 },
  { symbol: 'SPX', name: '标普500', price: 5864.67, change_pct: 0.45 },
  { symbol: 'NDX', name: '纳斯达克100', price: 20450.80, change_pct: 0.62 },
  { symbol: 'DJI', name: '道琼斯', price: 43800.00, change_pct: 0.28 },
  { symbol: 'US10Y', name: '美债10年收益率', price: 4.41, change_pct: -0.49 }
]

export const SEED_TW_MACRO = [
  { symbol: 'TWII', name: '台湾加权', price: 23500.00, change_pct: 1.15 },
  { symbol: 'TWOII', name: '台湾柜买', price: 268.50, change_pct: 0.82 },
  { symbol: 'TSMC', name: '台积电', price: 1050.00, change_pct: 1.45 }
]

export const SEED_CRYPTO_AGRI = [
  { symbol: 'BTC', name: '比特币', price: 95200.00, change_pct: -1.25 },
  { symbol: 'ETH', name: '以太坊', price: 3450.00, change_pct: -0.85 },
  { symbol: 'SOYBEAN', name: '美大豆', price: 1025.50, change_pct: 0.45 }
]

export const SEED_FX_LIST = [
  { symbol: 'CNYKRW', name: '人民币/韩元', price: '201.47', change_pct: -0.12, currency: 'KRW' },
  { symbol: 'CNYJPY', name: '人民币/日元', price: '23.73', change_pct: -1.02, currency: 'JPY' },
  { symbol: 'USDKRW', name: '美元/韩元', price: '1345.48', change_pct: -0.13, currency: 'KRW' },
  { symbol: 'USDJPY', name: '美元/日元', price: '158.53', change_pct: -0.81, currency: 'JPY' }
]

export const SEED_PRECIOUS_METALS = [
  { name: '黄金', price: 2712.50, change_pct: 1.05 },
  { name: '白银', price: 31.25, change_pct: 1.34 }
]

export const SEED_INDUSTRIAL_METALS = [
  { name: '铜', price: 9520.00, change_pct: 0.02 },
  { name: '铝', price: 2615.00, change_pct: -1.00 },
  { name: '锌', price: 2980.00, change_pct: 0.29 },
  { name: '镍', price: 16250.00, change_pct: -0.26 },
  { name: '锡', price: 31800.00, change_pct: 1.12 }
]

export const SEED_OTHER_METALS = [
  { name: '钨', price: 218.00, change_pct: -0.12 },
  { name: '钼', price: 3850.00, change_pct: -3.62 },
  { name: '锗', price: 18500.00, change_pct: -2.53 },
  { name: '铟', price: 2950.00, change_pct: -2.36 },
  { name: '锑', price: 158000.00, change_pct: -2.03 }
]

export const SEED_AI_PRODUCTS = [
  { name: '云算力', price: 3.85, change_pct: -4.99 },
  { name: 'Token', price: 0.015, change_pct: -1.82 }
]

export const SEED_AI_EQUIPMENTS = [
  { name: 'DRAM', price: 4.25, change_pct: -6.31 },
  { name: 'NAND', price: 2.85, change_pct: -4.76 },
  { name: 'HBM', price: 285.00, change_pct: -3.83 },
  { name: 'SSD', price: 68.50, change_pct: -4.63 },
  { name: '光模块', price: 450.00, change_pct: -4.89 },
  { name: '光纤', price: 12.80, change_pct: -3.66 },
  { name: 'PCB', price: 185.00, change_pct: -4.25 },
  { name: 'MLCC', price: 0.025, change_pct: -2.78 },
  { name: 'GPU', price: 32500.00, change_pct: -4.99 },
  { name: 'CPU', price: 1850.00, change_pct: -3.65 },
  { name: '先进制程', price: 18500.00, change_pct: -4.02 },
  { name: '封装', price: 65.00, change_pct: -3.83 },
  { name: '电力', price: 0.68, change_pct: 0.29 },
  { name: '电力设备', price: 101.04, change_pct: -0.86 },
  { name: '散热', price: 125.00, change_pct: 1.11 },
  { name: '算力租赁', price: 15.80, change_pct: -0.42 }
]

export const SEED_ROBOT_DRIVE = [
  { name: '行星滚柱丝杠', price: 1280.00, change_pct: 2.35 },
  { name: '谐波减速器', price: 860.00, change_pct: 3.12 },
  { name: 'RV减速器', price: 1650.00, change_pct: 1.45 },
  { name: '无框力矩电机', price: 450.00, change_pct: 2.80 },
  { name: '空心杯电机', price: 180.00, change_pct: 1.95 }
]

export const SEED_ROBOT_SENSE = [
  { name: '六维力矩传感器', price: 3200.00, change_pct: 4.10 },
  { name: '灵巧手', price: 18500.00, change_pct: 3.65 },
  { name: '触觉电子皮肤', price: 680.00, change_pct: 2.20 },
  { name: '3D视觉深度相机', price: 1250.00, change_pct: 1.80 },
  { name: '轻量化高强机体', price: 420.00, change_pct: 0.95 }
]

export const SEED_SEMI_EQUIP = [
  { name: '光刻设备', price: 35000.00, change_pct: 3.85 },
  { name: '等离子刻蚀机', price: 12500.00, change_pct: 4.20 },
  { name: '薄膜沉积PECVD', price: 9800.00, change_pct: 2.95 },
  { name: '离子注入机', price: 6800.00, change_pct: 2.15 },
  { name: '晶圆抛光CMP', price: 3400.00, change_pct: 1.80 }
]

export const SEED_SEMI_MATERIAL = [
  { name: '12英寸大硅片', price: 480.00, change_pct: 1.65 },
  { name: 'ArFi光刻胶', price: 1580.00, change_pct: 4.80 },
  { name: '电子级特种气体', price: 890.00, change_pct: 2.30 },
  { name: '先进封装CoWoS', price: 320.00, change_pct: 3.40 },
  { name: 'EDA设计软件', price: 1850.00, change_pct: 2.75 }
]

export const SEED_DIVIDEND_CORE = [
  { name: '大型水电', price: 103.80, change_pct: 0.85 },
  { name: '动力煤炭', price: 102.50, change_pct: 0.65 },
  { name: '石油管输', price: 101.90, change_pct: 0.40 },
  { name: '高速公路', price: 101.20, change_pct: 0.32 },
  { name: '港口航运', price: 102.10, change_pct: 0.70 },
  { name: '电信运营', price: 103.15, change_pct: 0.90 }
]

export const SEED_CYCLE_OVERVIEW: CycleOverview = {
  current_phase: '弱复苏与海外结构性繁荣分化',
  phase_code: 1,
  dominant_asset: '红利高股息与海外AI核心资产',
  equity_bond_spread: 2.8097,
  style_preference: '红利防守为主，低区伺机成长',
  summary: '国内十年期国债维持低位，红利股息率利差极具性价比；创业板50/红利低波100比值处于24%历史分位中间偏低区。'
}

export const SEED_STRATEGY_SIGNAL: CombinedSignal = {
  date: '2026-11-20',
  rotation: {
    date: '2026-11-20',
    ratio: 0.1846,
    zone: '中间区',
    quantile_pct: 24.33,
    action: '中间区：暂停指数买入，周定投资金1000元进现金池',
    to_high_pct: 51.67,
    to_low_pct: 13.33
  },
  scoring: {
    tech: 22.98,
    spread: 70.24,
    value: 9.23,
    style: 24.27,
    spread_pp: 2.81,
    total: 6.57,
    action: '空仓·暂停定投（老仓不动）',
    multiplier: 0.0,
    amount: 0.0,
    reduce_ratio: 0.0,
    close: 11878.27,
    pe: 9.61,
    pe_percentile: 0.9077,
    dividend_yield: 4.39,
    bond10y: 1.5803,
    rsi70: 53.57,
    b_pct100: 0.9620,
    style_percentile: 0.7573,
    data_source_note: '中证指数官网官方行情'
  },
  combined_action: '轮动中间区：红利打分仅作估值观察，不新增指数买入；资金1000元进入现金池等待',
  summary_tips: '当前比值在 0.16~0.28 中间区，维持现金池储备，耐受风格拉锯。'
}

export const SEED_DEMO_PORTFOLIO: PortfolioOverview = {
  account_id: 'ACC_DEMO_001',
  total_asset: 26850.0,
  available_cash: 5200.0,
  cash_pool: 1000.0,
  positions_value: 20650.0,
  total_cost: 20000.0,
  total_profit: 650.0,
  total_profit_pct: 3.25,
  positions: [
    { symbol: '930955', name: '红利低波100(515100)', shares: 12500, current_price: 1.1878, market_value: 14847.5, cost_price: 1.12, total_cost: 14000, unrealized_profit: 847.5, unrealized_profit_pct: 6.05 },
    { symbol: '399673', name: '创业板50(159949)', shares: 8000, current_price: 0.7253, market_value: 5802.4, cost_price: 0.75, total_cost: 6000, unrealized_profit: -197.6, unrealized_profit_pct: -3.29 }
  ],
  recent_transactions: [
    { date: '2026-11-16', event_type: 'cash_deposit', target_name: '计划现金池', amount: 1000, cash_pool_delta: 1000, cash_pool_after: 1000, description: '中间区：本周定投资金 1000 元存入计划现金池' },
    { date: '2026-11-09', event_type: 'weekly_invest', target_name: '红利低波100(515100)', amount: 1000, shares: 880, price: 1.136, description: '高区：周定投 1000 元买入红利低波100' }
  ]
}

export const FALLBACK_BACKTEST_DATA: BacktestOutput = {
  summary: {
    start_date: '2023-08-02',
    end_date: '2026-11-20',
    trading_days: 800,
    total_invest: 165000.0,
    final_asset: 186016.47,
    net_profit: 21016.47,
    total_return_pct: 12.74,
    annualized_return: 3.7,
    xirr: 7.39,
    max_drawdown: 24.94,
    sharpe_ratio: 2.42,
    calmar_ratio: 0.15,
    zone_distribution: {
      '低区': 163,
      '中间区': 435,
      '高区': 202
    },
    total_rebuy_count: 6,
    ledger_entries_count: 171
  },
  daily_records: [],
  ledger_entries: []
}
