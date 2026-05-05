import {
  Coupon,
  CouponType,
  DistributionRecord,
  UsedCouponInfo,
  ValidityType,
} from '@coupon/shared';

export function formatAmount(amount: number): string {
  const yuan = Math.floor(amount / 100);
  const fen = amount % 100;
  if (fen === 0) {
    return `¥${yuan}`;
  }
  return `¥${yuan}.${fen.toString().padStart(2, '0')}`;
}

export function formatCouponType(type: CouponType): string {
  switch (type) {
    case CouponType.FULL_REDUCTION:
      return '满减券';
    case CouponType.DISCOUNT:
      return '折扣券';
    case CouponType.FIXED_AMOUNT:
      return '固定金额券';
    default:
      return '未知类型';
  }
}

export function formatDiscountValue(type: CouponType, value: number): string {
  switch (type) {
    case CouponType.FULL_REDUCTION:
    case CouponType.FIXED_AMOUNT:
      return formatAmount(value);
    case CouponType.DISCOUNT:
      return `${value / 10}折`;
    default:
      return value.toString();
  }
}

export function formatValidity(validity: Coupon['validity']): string {
  if (validity.type === ValidityType.FIXED_DATE) {
    const start = new Date(validity.startDate).toLocaleString('zh-CN');
    const end = new Date(validity.endDate).toLocaleString('zh-CN');
    return `固定日期: ${start} 至 ${end}`;
  } else {
    return `领取后 ${validity.days} 天有效`;
  }
}

export function formatDate(isoString: string): string {
  return new Date(isoString).toLocaleString('zh-CN');
}

export function formatCoupon(coupon: Coupon): string {
  const lines: string[] = [];
  lines.push(`优惠券 ID: ${coupon.id}`);
  lines.push(`名称: ${coupon.name}`);
  lines.push(`类型: ${formatCouponType(coupon.type)}`);
  lines.push(`面值: ${formatDiscountValue(coupon.type, coupon.value)}`);
  lines.push(`使用门槛: ${formatAmount(coupon.threshold)}`);
  lines.push(`有效期: ${formatValidity(coupon.validity)}`);
  lines.push(`领取时间: ${formatDate(coupon.receiveTime)}`);
  lines.push(`过期时间: ${formatDate(coupon.expireTime)}`);
  lines.push(`状态: ${coupon.isUsed ? '已使用' : '未使用'}`);
  if (coupon.useTime) {
    lines.push(`使用时间: ${formatDate(coupon.useTime)}`);
  }
  return lines.join('\n');
}

export function formatCoupons(coupons: Coupon[]): string {
  if (coupons.length === 0) {
    return '没有找到优惠券';
  }

  const lines: string[] = [];
  lines.push(`共找到 ${coupons.length} 张优惠券：`);
  lines.push('');

  coupons.forEach((coupon, index) => {
    lines.push(`[${index + 1}]`);
    lines.push(formatCoupon(coupon));
    lines.push('');
  });

  return lines.join('\n');
}

export function formatDistributionRecords(records: DistributionRecord[]): string {
  if (records.length === 0) {
    return '没有发放记录';
  }

  const lines: string[] = [];
  lines.push(`成功发放 ${records.length} 张优惠券：`);
  lines.push('');

  records.forEach((record, index) => {
    lines.push(`[${index + 1}]`);
    lines.push(`发放记录 ID: ${record.id}`);
    lines.push(`发放人: ${record.distributorId}`);
    lines.push(`领取人: ${record.receiveUserId}`);
    lines.push(`优惠券 ID: ${record.couponId}`);
    lines.push(`发放时间: ${formatDate(record.distributionTime)}`);
    lines.push('');
  });

  return lines.join('\n');
}

export function formatRedeemResult(
  originalAmount: number,
  discountAmount: number,
  finalAmount: number,
  usedCoupons: UsedCouponInfo[]
): string {
  const lines: string[] = [];
  lines.push('=== 核销结果 ===');
  lines.push(`原价: ${formatAmount(originalAmount)}`);
  lines.push(`优惠: ${formatAmount(discountAmount)}`);
  lines.push(`实付: ${formatAmount(finalAmount)}`);

  if (usedCoupons.length > 0) {
    lines.push('');
    lines.push('使用的优惠券:');
    usedCoupons.forEach((coupon, index) => {
      lines.push(`[${index + 1}] ${coupon.name} (${formatCouponType(coupon.type)}) - 优惠 ${formatAmount(coupon.discountAmount)}`);
    });
  }

  return lines.join('\n');
}

export function formatRecommendResult(
  recommendedCouponIds: string[],
  totalDiscount: number
): string {
  if (recommendedCouponIds.length === 0) {
    return '没有找到合适的优惠券';
  }

  const lines: string[] = [];
  lines.push('=== 推荐优惠券 ===');
  lines.push(`预计优惠: ${formatAmount(totalDiscount)}`);
  lines.push('推荐使用的优惠券 ID:');
  recommendedCouponIds.forEach((id, index) => {
    lines.push(`  ${index + 1}. ${id}`);
  });

  return lines.join('\n');
}

export function formatError(errorCode?: number, errorMessage?: string): string {
  const lines: string[] = [];
  lines.push('错误:');
  if (errorCode !== undefined) {
    lines.push(`错误码: ${errorCode}`);
  }
  if (errorMessage) {
    lines.push(`错误信息: ${errorMessage}`);
  }
  return lines.join('\n');
}

export function printHelp(): void {
  const helpText = `
优惠券发放和核销 CLI 工具

用法: coupon <command> [options]

命令:
  distribute [options]    发放优惠券
  redeem [options]        核销优惠券
  query [options]         查询用户优惠券
  recommend [options]     获取推荐优惠券
  help                    显示此帮助信息

发放优惠券 (distribute) 选项:
  --type <type>           优惠券类型: FULL_REDUCTION, DISCOUNT, FIXED_AMOUNT (必填)
  --name <name>           优惠券名称 (必填)
  --value <number>        面值 (分, 折扣券为 1-100 的整数, 50 表示 5 折) (必填)
  --threshold <number>    使用门槛 (分, 0 表示无门槛) (必填)
  --validity-type <type>  有效期类型: FIXED_DATE, DAYS_AFTER_RECEIVE (必填)
  --start-date <date>     有效期开始日期 (FIXED_DATE 时必填, ISO 格式)
  --end-date <date>       有效期结束日期 (FIXED_DATE 时必填, ISO 格式)
  --days <number>         领取后多少天有效 (DAYS_AFTER_RECEIVE 时必填)
  --distribution-type <type>  发放类型: TARGETED (定向), ALL (全员) (必填)
  --user-ids <ids>        用户 ID 列表, 逗号分隔 (TARGETED 时必填)
  --user-filter <filter>  用户筛选: ALL, NEW_USERS, OLD_USERS (ALL 发放时可选)
  --new-user-days <days>  新用户阈值天数, 默认 7 天 (NEW_USERS/OLD_USERS 时可选)
  --distributor-id <id>   发放人 ID (必填)

核销优惠券 (redeem) 选项:
  --user-id <id>          用户 ID (必填)
  --order-amount <number> 订单金额 (分) (必填)
  --coupon-ids <ids>      优惠券 ID 列表, 逗号分隔 (可选, 不指定则使用推荐)

查询用户优惠券 (query) 选项:
  --user-id <id>          用户 ID (必填)
  --include-used          包含已使用的优惠券 (可选)
  --include-expired       包含已过期的优惠券 (可选)

获取推荐优惠券 (recommend) 选项:
  --user-id <id>          用户 ID (必填)
  --order-amount <number> 订单金额 (分) (必填)

示例:
  # 发放满减券给指定用户
  coupon distribute --type FULL_REDUCTION --name "满100减20" --value 2000 --threshold 10000 \
    --validity-type DAYS_AFTER_RECEIVE --days 7 \
    --distribution-type TARGETED --user-ids "user-001,user-002" --distributor-id "admin-001"

  # 发放折扣券给所有新用户 (注册 7 天内)
  coupon distribute --type DISCOUNT --name "新用户8折" --value 80 --threshold 0 \
    --validity-type FIXED_DATE --start-date "2024-01-01T00:00:00Z" --end-date "2024-12-31T23:59:59Z" \
    --distribution-type ALL --user-filter NEW_USERS --distributor-id "admin-001"

  # 核销优惠券
  coupon redeem --user-id "user-001" --order-amount 50000 --coupon-ids "coupon-id-1,coupon-id-2"

  # 查询用户优惠券
  coupon query --user-id "user-001" --include-used --include-expired

  # 获取推荐优惠券
  coupon recommend --user-id "user-001" --order-amount 50000
`;
  console.log(helpText);
}
