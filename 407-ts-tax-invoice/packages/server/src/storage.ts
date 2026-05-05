import fs from "node:fs/promises";
import path from "node:path";
import { Invoice, InvoiceId } from "@tax-invoice/shared";
import { config, getDataFilePath } from "./config.js";

export interface InvoiceStore {
  getById(id: InvoiceId): Invoice | undefined;
  getByInvoiceNumber(invoiceNumber: string): Invoice | undefined;
  getAll(): Invoice[];
  save(invoice: Invoice): void;
  update(invoice: Invoice): void;
  delete(id: InvoiceId): boolean;
  saveToFile(): Promise<void>;
  loadFromFile(): Promise<void>;
  startAutoSave(): void;
  stopAutoSave(): void;
}

export class MemoryInvoiceStore implements InvoiceStore {
  private invoices: Map<InvoiceId, Invoice> = new Map();
  private invoiceNumberToId: Map<string, InvoiceId> = new Map();
  private autoSaveTimer: NodeJS.Timeout | null = null;

  getById(id: InvoiceId): Invoice | undefined {
    return this.invoices.get(id);
  }

  getByInvoiceNumber(invoiceNumber: string): Invoice | undefined {
    const id = this.invoiceNumberToId.get(invoiceNumber);
    return id ? this.invoices.get(id) : undefined;
  }

  getAll(): Invoice[] {
    return Array.from(this.invoices.values());
  }

  save(invoice: Invoice): void {
    this.invoices.set(invoice.id, invoice);
    this.invoiceNumberToId.set(invoice.invoiceNumber, invoice.id);
  }

  update(invoice: Invoice): void {
    if (!this.invoices.has(invoice.id)) {
      throw new Error(`Invoice not found: ${invoice.id}`);
    }
    this.invoices.set(invoice.id, invoice);
  }

  delete(id: InvoiceId): boolean {
    const invoice = this.invoices.get(id);
    if (invoice) {
      this.invoiceNumberToId.delete(invoice.invoiceNumber);
    }
    return this.invoices.delete(id);
  }

  async saveToFile(): Promise<void> {
    const dataPath = getDataFilePath();
    const dirPath = path.dirname(dataPath);

    try {
      await fs.mkdir(dirPath, { recursive: true });
      const invoices = this.getAll();
      const jsonContent = JSON.stringify(invoices, null, 2);
      await fs.writeFile(dataPath, jsonContent, "utf-8");
    } catch (error) {
      console.error("Failed to save invoices to file:", error);
      throw error;
    }
  }

  async loadFromFile(): Promise<void> {
    const dataPath = getDataFilePath();

    try {
      const fileContent = await fs.readFile(dataPath, "utf-8");
      const invoices: Invoice[] = JSON.parse(fileContent);

      this.invoices.clear();
      this.invoiceNumberToId.clear();

      for (const invoice of invoices) {
        this.invoices.set(invoice.id, invoice);
        this.invoiceNumberToId.set(invoice.invoiceNumber, invoice.id);
      }

      console.log(`Loaded ${invoices.length} invoices from file`);
    } catch (error) {
      if (error instanceof Error && "code" in error && error.code === "ENOENT") {
        console.log("No data file found, starting with empty store");
      } else {
        console.error("Failed to load invoices from file:", error);
        throw error;
      }
    }
  }

  startAutoSave(): void {
    if (this.autoSaveTimer) {
      return;
    }

    this.autoSaveTimer = setInterval(async () => {
      try {
        await this.saveToFile();
      } catch (error) {
        console.error("Auto-save failed:", error);
      }
    }, config.saveInterval);

    console.log(`Auto-save enabled (interval: ${config.saveInterval}ms)`);
  }

  stopAutoSave(): void {
    if (this.autoSaveTimer) {
      clearInterval(this.autoSaveTimer);
      this.autoSaveTimer = null;
      console.log("Auto-save disabled");
    }
  }
}

export const store = new MemoryInvoiceStore();
