import {
  Invoice,
  InvoiceStatus,
  InvoiceType,
  MonthlySummary,
  BatchImportResult,
} from "@tax-invoice/shared";
import { formatAmountToYuan } from "@tax-invoice/shared";

function formatInvoiceType(type: InvoiceType): string {
  return type === InvoiceType.SPECIAL ? "专票" : "普票";
}

function formatInvoiceStatus(status: InvoiceStatus): string {
  switch (status) {
    case InvoiceStatus.NORMAL:
      return "正常";
    case InvoiceStatus.VOIDED:
      return "已作废";
    case InvoiceStatus.RED_INVOICED:
      return "已红冲";
  }
}

export function formatInvoice(invoice: Invoice, indent = 0): string {
  const prefix = "  ".repeat(indent);
  const lines: string[] = [];

  lines.push(`${prefix}ID: ${invoice.id}`);
  lines.push(`${prefix}发票号码: ${invoice.invoiceNumber}`);
  lines.push(`${prefix}发票类型: ${formatInvoiceType(invoice.invoiceType)}`);
  lines.push(`${prefix}状态: ${formatInvoiceStatus(invoice.status)}`);
  lines.push("");
  lines.push(`${prefix}购买方:`);
  lines.push(`${prefix}  名称: ${invoice.buyer.name}`);
  lines.push(`${prefix}  税号: ${invoice.buyer.taxNumber}`);
  lines.push("");
  lines.push(`${prefix}销售方:`);
  lines.push(`${prefix}  名称: ${invoice.seller.name}`);
  lines.push(`${prefix}  税号: ${invoice.seller.taxNumber}`);
  lines.push("");
  lines.push(`${prefix}金额信息:`);
  lines.push(`${prefix}  金额不含税: ${formatAmountToYuan(invoice.amountExcludingTax)} 元`);
  lines.push(`${prefix}  税额: ${formatAmountToYuan(invoice.taxAmount)} 元`);
  lines.push(`${prefix}  价税合计: ${formatAmountToYuan(invoice.totalAmount)} 元`);
  lines.push("");
  lines.push(`${prefix}抵扣状态: ${invoice.isDeducted ? "已抵扣" : "未抵扣"}`);

  if (invoice.reimbursementId) {
    lines.push(`${prefix}关联报销单: ${invoice.reimbursementId}`);
  }
  if (invoice.redInvoiceId) {
    lines.push(`${prefix}红字发票ID: ${invoice.redInvoiceId}`);
  }
  if (invoice.originalInvoiceId) {
    lines.push(`${prefix}原发票ID: ${invoice.originalInvoiceId}`);
  }

  lines.push("");
  lines.push(`${prefix}开票日期: ${new Date(invoice.issuedAt).toLocaleString()}`);
  lines.push(`${prefix}创建时间: ${new Date(invoice.createdAt).toLocaleString()}`);
  lines.push(`${prefix}更新时间: ${new Date(invoice.updatedAt).toLocaleString()}`);

  return lines.join("\n");
}

export function formatInvoiceList(invoices: Invoice[]): string {
  if (invoices.length === 0) {
    return "没有找到发票记录";
  }

  const lines: string[] = [];
  lines.push(`共找到 ${invoices.length} 条发票记录:`);
  lines.push("");

  invoices.forEach((invoice, index) => {
    lines.push(`--- 发票 ${index + 1} ---`);
    lines.push(formatInvoice(invoice));
    if (index < invoices.length - 1) {
      lines.push("");
    }
  });

  return lines.join("\n");
}

export function formatMonthlySummary(summary: MonthlySummary): string {
  const lines: string[] = [];

  lines.push(`=== ${summary.year}年${summary.month}月 发票汇总 ===`);
  lines.push("");

  lines.push("【增值税专用发票】");
  lines.push(`  数量: ${summary.specialInvoice.count} 张`);
  lines.push(`  金额不含税: ${formatAmountToYuan(summary.specialInvoice.totalAmountExcludingTax)} 元`);
  lines.push(`  税额: ${formatAmountToYuan(summary.specialInvoice.totalTaxAmount)} 元`);
  lines.push(`  价税合计: ${formatAmountToYuan(summary.specialInvoice.totalAmount)} 元`);
  lines.push("");

  lines.push("【增值税普通发票】");
  lines.push(`  数量: ${summary.normalInvoice.count} 张`);
  lines.push(`  金额不含税: ${formatAmountToYuan(summary.normalInvoice.totalAmountExcludingTax)} 元`);
  lines.push(`  税额: ${formatAmountToYuan(summary.normalInvoice.totalTaxAmount)} 元`);
  lines.push(`  价税合计: ${formatAmountToYuan(summary.normalInvoice.totalAmount)} 元`);
  lines.push("");

  const totalCount = summary.specialInvoice.count + summary.normalInvoice.count;
  const totalAmountExcludingTax = summary.specialInvoice.totalAmountExcludingTax + summary.normalInvoice.totalAmountExcludingTax;
  const totalTaxAmount = summary.specialInvoice.totalTaxAmount + summary.normalInvoice.totalTaxAmount;
  const totalAmount = summary.specialInvoice.totalAmount + summary.normalInvoice.totalAmount;

  lines.push("【合计】");
  lines.push(`  数量: ${totalCount} 张`);
  lines.push(`  金额不含税: ${formatAmountToYuan(totalAmountExcludingTax)} 元`);
  lines.push(`  税额: ${formatAmountToYuan(totalTaxAmount)} 元`);
  lines.push(`  价税合计: ${formatAmountToYuan(totalAmount)} 元`);

  return lines.join("\n");
}

export function formatBatchImportResult(result: BatchImportResult): string {
  const lines: string[] = [];

  lines.push("=== 批量导入结果 ===");
  lines.push(`成功: ${result.success} 张`);
  lines.push(`失败: ${result.failed} 张`);
  lines.push("");

  if (result.successItems.length > 0) {
    lines.push("【成功导入的发票】");
    result.successItems.forEach((invoice, index) => {
      lines.push(`  ${index + 1}. ${invoice.invoiceNumber} (ID: ${invoice.id})`);
    });
    lines.push("");
  }

  if (result.failedItems.length > 0) {
    lines.push("【导入失败的发票】");
    result.failedItems.forEach((item) => {
      lines.push(`  索引 ${item.index}: ${item.data.invoiceNumber || "未知"}`);
      lines.push(`    原因: ${item.reason}`);
    });
  }

  return lines.join("\n");
}

export function formatError(code: string, message: string, details?: unknown): string {
  const lines: string[] = [];

  lines.push("❌ 错误");
  lines.push(`  错误码: ${code}`);
  lines.push(`  消息: ${message}`);

  if (details) {
    lines.push(`  详情: ${JSON.stringify(details, null, 2)}`);
  }

  return lines.join("\n");
}

export function formatSuccess(message: string, data?: unknown): string {
  const lines: string[] = [];

  lines.push(`✅ ${message}`);

  if (data !== undefined && data !== null) {
    if (typeof data === "object") {
      lines.push(JSON.stringify(data, null, 2));
    } else {
      lines.push(String(data));
    }
  }

  return lines.join("\n");
}
