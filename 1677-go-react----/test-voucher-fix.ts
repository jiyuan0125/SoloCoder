import { v4 as uuidv4 } from 'uuid';
import { db } from './src/database';
import dayjs from 'dayjs';

function testVoucherLogic() {
  console.log('=== 测试税前扣除凭证限额计算逻辑 ===\n');
  
  const testProjectId = uuidv4();
  const testDonationId = uuidv4();
  const donorName = '测试用户';
  const idCardLast4 = '1234';
  const donationAmount = 500000;
  const taxIncomeBase = 10000000;
  const taxYear = dayjs().year();
  
  console.log('测试数据:');
  console.log(`  - 捐赠金额: ${donationAmount} 分 (${donationAmount / 100} 元)`);
  console.log(`  - 应纳税所得额: ${taxIncomeBase} 分 (${taxIncomeBase / 100} 元)`);
  console.log(`  - 30% 限额: ${taxIncomeBase * 0.3} 分 (${taxIncomeBase * 0.3 / 100} 元)`);
  console.log(`  - 捐赠人: ${donorName} (${idCardLast4})`);
  console.log();
  
  const limit = Math.floor(taxIncomeBase * 0.3);
  console.log(`计算限额: ${limit} 分 (${limit / 100} 元)`);
  console.log(`捐赠金额 ${donationAmount} < 限额 ${limit} ? ${donationAmount < limit}`);
  console.log();
  
  const remainingLimit = limit - 0;
  const deductibleAmount = Math.min(donationAmount, remainingLimit);
  console.log(`可抵扣金额: ${deductibleAmount} 分 (${deductibleAmount / 100} 元)`);
  console.log(`deductibleAmount <= 0 ? ${deductibleAmount <= 0}`);
  console.log();
  
  if (deductibleAmount > 0) {
    console.log('✅ 凭证应该可以正常生成！');
  } else {
    console.log('❌ 凭证生成失败（已达限额）');
  }
  
  console.log();
  console.log('=== 检查数据库中的现有记录 ===');
  
  const existingSummary = db.prepare(
    `SELECT * FROM donor_annual_summary 
     WHERE donor_name = ? AND id_card_last4 = ? AND year = ?`
  ).get(donorName, idCardLast4, taxYear);
  
  if (existingSummary) {
    console.log('找到现有记录:');
    console.log('  ', existingSummary);
    console.log();
    console.log(`  tax_income_base = ${existingSummary.tax_income_base}`);
    console.log(`  tax_limit = ${existingSummary.tax_limit}`);
    console.log(`  total_tax_deductible_amount = ${existingSummary.total_tax_deductible_amount}`);
    
    const effectiveIncomeBase = Number(existingSummary.tax_income_base) || 0;
    const storedLimit = Number(existingSummary.tax_limit) || 0;
    const currentTotal = Number(existingSummary.total_tax_deductible_amount) || 0;
    
    console.log();
    console.log('使用修复后的逻辑重新计算:');
    console.log(`  effectiveIncomeBase = ${effectiveIncomeBase}`);
    console.log(`  newTaxIncomeBase = ${taxIncomeBase}`);
    
    let taxLimit = storedLimit;
    if (taxIncomeBase > effectiveIncomeBase) {
      taxLimit = Math.floor(taxIncomeBase * 0.3);
      console.log(`  新收入基数更大，更新限额为: ${taxLimit}`);
    } else if (effectiveIncomeBase > 0) {
      taxLimit = Math.floor(effectiveIncomeBase * 0.3);
      console.log(`  使用现有收入基数重新计算限额: ${taxLimit}`);
    }
    
    console.log(`  currentTotal = ${currentTotal}`);
    console.log(`  limit = ${taxLimit}`);
    console.log(`  currentTotal >= limit ? ${currentTotal >= taxLimit}`);
    
    const newRemainingLimit = taxLimit - currentTotal;
    const newDeductibleAmount = Math.min(donationAmount, newRemainingLimit);
    console.log(`  可抵扣金额: ${newDeductibleAmount}`);
    
    if (newDeductibleAmount > 0) {
      console.log('\n✅ 修复后，凭证应该可以正常生成！');
    } else {
      console.log('\n❌ 修复后仍然失败');
    }
  } else {
    console.log('未找到现有记录（这是第一次操作）');
    console.log('创建新记录时:');
    console.log(`  tax_income_base = ${taxIncomeBase}`);
    console.log(`  tax_limit = ${limit}`);
    console.log(`  currentTotal = 0`);
    console.log(`  deductibleAmount = ${deductibleAmount}`);
    console.log('\n✅ 凭证应该可以正常生成！');
  }
}

testVoucherLogic();
