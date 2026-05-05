import {
  Invoice,
  CreateInvoiceRequest,
  InvoiceType,
  InvoiceFilter,
  UpdateInvoiceRequest,
  BatchImportResult,
  MonthlySummary,
} from "@tax-invoice/shared";
import {
  createInvoice,
  getInvoice,
  queryInvoices,
  updateInvoice,
  voidInvoice,
  redInvoice,
  batchImport,
  getMonthlySummary,
} from "./client.js";
import {
  formatInvoice,
  formatInvoiceList,
  formatMonthlySummary,
  formatBatchImportResult,
  formatError,
  formatSuccess,
} from "./format.js";

interface CommandArgs {
  [key: string]: string | number | boolean | undefined;
}

function parseArgs(args: string[]): { command: string; flags: CommandArgs } {
  const command = args[0] || "help";
  const flags: CommandArgs = {};

  for (let i = 1; i < args.length; i++) {
    const arg = args[i];
    if (!arg) continue;

    if (arg.startsWith("--")) {
      const parts = arg.substring(2).split("=");
      const key = parts[0];
      const value = parts.length > 1 ? parts.slice(1).join("=") : undefined;

      if (!key) continue;

      if (value === undefined) {
        flags[key] = true;
      } else if (value === "true") {
        flags[key] = true;
      } else if (value === "false") {
        flags[key] = false;
      } else if (!isNaN(Number(value)) && value !== "") {
        flags[key] = Number(value);
      } else {
        flags[key] = value;
      }
    }
  }

  return { command, flags };
}

function parseInvoiceType(type: string): InvoiceType {
  if (type === "special" || type === "专票") {
    return InvoiceType.SPECIAL;
  }
  if (type === "normal" || type === "普票") {
    return InvoiceType.NORMAL;
  }
  throw new Error(`无效的发票类型: ${type}，请使用 special/专票 或 normal/普票`);
}

export async function runCommand(args: string[]): Promise<void> {
  const { command, flags } = parseArgs(args);

  switch (command) {
    case "create":
    case "新增":
      await handleCreate(flags);
      break;

    case "get":
    case "查询":
      await handleGet(flags);
      break;

    case "list":
    case "列表":
      await handleList(flags);
      break;

    case "update":
    case "更新":
      await handleUpdate(flags);
      break;

    case "void":
    case "作废":
      await handleVoid(flags);
      break;

    case "red":
    case "红冲":
      await handleRed(flags);
      break;

    case "batch":
    case "批量导入":
      await handleBatch(flags);
      break;

    case "summary":
    case "汇总":
      await handleSummary(flags);
      break;

    case "help":
    case "--help":
    case "-h":
    default:
      showHelp();
      break;
  }
}

async function handleCreate(flags: CommandArgs): Promise<void> {
  const required = ["invoiceNumber", "invoiceType", "buyerName", "buyerTaxNumber", "sellerName", "sellerTaxNumber", "amountExcludingTax", "taxAmount", "totalAmount", "issuedAt"];
  const missing = required.filter((key) => flags[key] === undefined);

  if (missing.length > 0) {
    console.error(formatError("MISSING_ARGS", `缺少必需参数: ${missing.join(", ")}`));
    console.log("\n用法: tax-invoice create [参数]");
    console.log("参数:");
    console.log("  --invoiceNumber=<发票号码>");
    console.log("  --invoiceType=<special|normal|专票|普票>");
    console.log("  --buyerName=<购买方名称>");
    console.log("  --buyerTaxNumber=<购买方税号>");
    console.log("  --sellerName=<销售方名称>");
    console.log("  --sellerTaxNumber=<销售方税号>");
    console.log("  --amountExcludingTax=<金额不含税(分)>");
    console.log("  --taxAmount=<税额(分)>");
    console.log("  --totalAmount=<价税合计(分)>");
    console.log("  --isDeducted=<true|false> (默认 false)");
    console.log("  --issuedAt=<开票日期 (ISO格式)>");
    return;
  }

  try {
    const request: CreateInvoiceRequest = {
      invoiceNumber: String(flags.invoiceNumber),
      invoiceType: parseInvoiceType(String(flags.invoiceType)),
      buyer: {
        name: String(flags.buyerName),
        taxNumber: String(flags.buyerTaxNumber),
      },
      seller: {
        name: String(flags.sellerName),
        taxNumber: String(flags.sellerTaxNumber),
      },
      amountExcludingTax: Number(flags.amountExcludingTax),
      taxAmount: Number(flags.taxAmount),
      totalAmount: Number(flags.totalAmount),
      isDeducted: Boolean(flags.isDeducted),
      issuedAt: String(flags.issuedAt),
    };

    const response = await createInvoice(request);

    if (response.success && response.data) {
      console.log(formatSuccess("发票创建成功"));
      console.log("");
      console.log(formatInvoice(response.data as Invoice));
    } else {
      console.error(formatError(response.error?.code || "UNKNOWN", response.error?.message || "未知错误", response.error?.details));
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    console.error(formatError("CONNECTION_ERROR", `连接服务器失败: ${message}`));
  }
}

async function handleGet(flags: CommandArgs): Promise<void> {
  if (!flags.id) {
    console.error(formatError("MISSING_ARGS", "缺少发票ID"));
    console.log("\n用法: tax-invoice get --id=<发票ID>");
    return;
  }

  try {
    const response = await getInvoice(String(flags.id));

    if (response.success && response.data) {
      console.log(formatInvoice(response.data as Invoice));
    } else {
      console.error(formatError(response.error?.code || "UNKNOWN", response.error?.message || "未知错误", response.error?.details));
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    console.error(formatError("CONNECTION_ERROR", `连接服务器失败: ${message}`));
  }
}

async function handleList(flags: CommandArgs): Promise<void> {
  try {
    const filter: InvoiceFilter = {};

    if (flags.taxAmountMin !== undefined) {
      filter.taxAmountMin = Number(flags.taxAmountMin);
    }
    if (flags.taxAmountMax !== undefined) {
      filter.taxAmountMax = Number(flags.taxAmountMax);
    }
    if (flags.invoiceType !== undefined) {
      filter.invoiceType = parseInvoiceType(String(flags.invoiceType));
    }
    if (flags.status !== undefined) {
      filter.status = String(flags.status) as InvoiceFilter["status"];
    }
    if (flags.startDate !== undefined) {
      filter.startDate = String(flags.startDate);
    }
    if (flags.endDate !== undefined) {
      filter.endDate = String(flags.endDate);
    }

    const response = await queryInvoices(filter);

    if (response.success && response.data) {
      const data = response.data as { items: Invoice[]; total: number };
      console.log(formatInvoiceList(data.items));
    } else {
      console.error(formatError(response.error?.code || "UNKNOWN", response.error?.message || "未知错误", response.error?.details));
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    console.error(formatError("CONNECTION_ERROR", `连接服务器失败: ${message}`));
  }
}

async function handleUpdate(flags: CommandArgs): Promise<void> {
  if (!flags.id) {
    console.error(formatError("MISSING_ARGS", "缺少发票ID"));
    console.log("\n用法: tax-invoice update --id=<发票ID> [--reimbursementId=<报销单ID>] [--isDeducted=<true|false>]");
    return;
  }

  try {
    const request: UpdateInvoiceRequest = {
      id: String(flags.id),
    };

    if (flags.reimbursementId !== undefined) {
      request.reimbursementId = String(flags.reimbursementId) || undefined;
    }
    if (flags.isDeducted !== undefined) {
      request.isDeducted = Boolean(flags.isDeducted);
    }

    const response = await updateInvoice(request);

    if (response.success && response.data) {
      console.log(formatSuccess("发票更新成功"));
      console.log("");
      console.log(formatInvoice(response.data as Invoice));
    } else {
      console.error(formatError(response.error?.code || "UNKNOWN", response.error?.message || "未知错误", response.error?.details));
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    console.error(formatError("CONNECTION_ERROR", `连接服务器失败: ${message}`));
  }
}

async function handleVoid(flags: CommandArgs): Promise<void> {
  if (!flags.id) {
    console.error(formatError("MISSING_ARGS", "缺少发票ID"));
    console.log("\n用法: tax-invoice void --id=<发票ID>");
    console.log("说明: 作废操作适用于未抵扣的发票");
    return;
  }

  try {
    const response = await voidInvoice(String(flags.id));

    if (response.success && response.data) {
      console.log(formatSuccess("发票作废成功"));
      console.log("");
      console.log(formatInvoice(response.data as Invoice));
    } else {
      console.error(formatError(response.error?.code || "UNKNOWN", response.error?.message || "未知错误", response.error?.details));
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    console.error(formatError("CONNECTION_ERROR", `连接服务器失败: ${message}`));
  }
}

async function handleRed(flags: CommandArgs): Promise<void> {
  const required = ["id", "invoiceNumber", "buyerName", "buyerTaxNumber", "sellerName", "sellerTaxNumber", "issuedAt"];
  const missing = required.filter((key) => flags[key] === undefined);

  if (missing.length > 0) {
    console.error(formatError("MISSING_ARGS", `缺少必需参数: ${missing.join(", ")}`));
    console.log("\n用法: tax-invoice red [参数]");
    console.log("参数:");
    console.log("  --id=<原发票ID>");
    console.log("  --invoiceNumber=<红字发票号码>");
    console.log("  --buyerName=<购买方名称>");
    console.log("  --buyerTaxNumber=<购买方税号>");
    console.log("  --sellerName=<销售方名称>");
    console.log("  --sellerTaxNumber=<销售方税号>");
    console.log("  --issuedAt=<开票日期 (ISO格式)>");
    console.log("说明: 红冲操作适用于已抵扣的发票，红字发票金额将自动与原发票一致");
    return;
  }

  try {
    const response = await redInvoice(String(flags.id), {
      invoiceNumber: String(flags.invoiceNumber),
      buyer: {
        name: String(flags.buyerName),
        taxNumber: String(flags.buyerTaxNumber),
      },
      seller: {
        name: String(flags.sellerName),
        taxNumber: String(flags.sellerTaxNumber),
      },
      issuedAt: String(flags.issuedAt),
    });

    if (response.success && response.data) {
      const data = response.data as { originalInvoice: Invoice; redInvoice: Invoice };
      console.log(formatSuccess("发票红冲成功"));
      console.log("");
      console.log("=== 原发票 ===");
      console.log(formatInvoice(data.originalInvoice));
      console.log("");
      console.log("=== 红字发票 ===");
      console.log(formatInvoice(data.redInvoice));
    } else {
      console.error(formatError(response.error?.code || "UNKNOWN", response.error?.message || "未知错误", response.error?.details));
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    console.error(formatError("CONNECTION_ERROR", `连接服务器失败: ${message}`));
  }
}

async function handleBatch(flags: CommandArgs): Promise<void> {
  if (!flags.file) {
    console.error(formatError("MISSING_ARGS", "缺少发票数据文件路径"));
    console.log("\n用法: tax-invoice batch --file=<JSON文件路径>");
    console.log("说明: JSON文件应包含一个发票数组，格式与 create 命令相同");
    return;
  }

  try {
    const { default: fs } = await import("node:fs/promises");
    const fileContent = await fs.readFile(String(flags.file), "utf-8");
    const invoices: CreateInvoiceRequest[] = JSON.parse(fileContent);

    const response = await batchImport(invoices);

    if (response.success && response.data) {
      console.log(formatBatchImportResult(response.data as BatchImportResult));
    } else {
      console.error(formatError(response.error?.code || "UNKNOWN", response.error?.message || "未知错误", response.error?.details));
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    console.error(formatError("ERROR", `批量导入失败: ${message}`));
  }
}

async function handleSummary(flags: CommandArgs): Promise<void> {
  try {
    let year: number | undefined;
    let month: number | undefined;

    if (flags.year !== undefined) {
      year = Number(flags.year);
    }
    if (flags.month !== undefined) {
      month = Number(flags.month);
    }

    const response = await getMonthlySummary(year, month);

    if (response.success && response.data) {
      console.log(formatMonthlySummary(response.data as MonthlySummary));
    } else {
      console.error(formatError(response.error?.code || "UNKNOWN", response.error?.message || "未知错误", response.error?.details));
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    console.error(formatError("CONNECTION_ERROR", `连接服务器失败: ${message}`));
  }
}

function showHelp(): void {
  const helpText = `
发票管理 CLI 工具

用法: tax-invoice <命令> [参数]

命令:
  create, 新增     创建新发票
  get, 查询        根据ID查询发票
  list, 列表       查询发票列表 (支持筛选)
  update, 更新     更新发票信息
  void, 作废       作废发票 (未抵扣时)
  red, 红冲        红冲发票 (已抵扣时)
  batch, 批量导入  批量导入发票
  summary, 汇总    月度汇总统计
  help, -h, --help 显示此帮助信息

示例:
  # 创建发票
  tax-invoice create --invoiceNumber=INV001 --invoiceType=special \\
    --buyerName="公司A" --buyerTaxNumber=91110000MA001G5R9A \\
    --sellerName="公司B" --sellerTaxNumber=91110000MA002G6R0B \\
    --amountExcludingTax=10000 --taxAmount=1300 --totalAmount=11300 \\
    --isDeducted=false --issuedAt=2024-01-15T00:00:00.000Z

  # 查询发票列表
  tax-invoice list --taxAmountMin=1000 --taxAmountMax=5000

  # 作废发票
  tax-invoice void --id=INV-ABC123

  # 红冲发票
  tax-invoice red --id=INV-ABC123 --invoiceNumber=RED001 \\
    --buyerName="公司A" --buyerTaxNumber=91110000MA001G5R9A \\
    --sellerName="公司B" --sellerTaxNumber=91110000MA002G6R0B \\
    --issuedAt=2024-01-20T00:00:00.000Z

  # 月度汇总
  tax-invoice summary --year=2024 --month=1

环境变量:
  TAX_INVOICE_API_URL  服务端地址 (默认: http://127.0.0.1:3000)

注意:
  - 所有金额字段以分为单位的整数
  - 日期使用 ISO 格式 (如: 2024-01-15T00:00:00.000Z)
  - 作废和红冲是不同操作，不能混用
`;
  console.log(helpText);
}
