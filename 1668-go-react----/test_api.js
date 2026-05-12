const http = require('http');

function request(method, path, body = null) {
  return new Promise((resolve, reject) => {
    const data = body ? JSON.stringify(body) : null;
    const options = {
      hostname: 'localhost',
      port: 3000,
      path,
      method,
      headers: {
        'Content-Type': 'application/json',
        ...(data ? { 'Content-Length': Buffer.byteLength(data) } : {})
      }
    };

    const req = http.request(options, (res) => {
      let responseBody = '';
      res.on('data', (chunk) => {
        responseBody += chunk;
      });
      res.on('end', () => {
        try {
          resolve({
            statusCode: res.statusCode,
            body: responseBody ? JSON.parse(responseBody) : null
          });
        } catch (e) {
          resolve({
            statusCode: res.statusCode,
            body: responseBody
          });
        }
      });
    });

    req.on('error', reject);
    if (data) req.write(data);
    req.end();
  });
}

async function runTests() {
  console.log('=== 发票管理系统 API 测试 ===\n');

  console.log('1. 健康检查...');
  const health = await request('GET', '/health');
  console.log(`   状态码: ${health.statusCode}, 响应: ${JSON.stringify(health.body)}`);
  console.assert(health.statusCode === 200, '健康检查应该返回 200');

  console.log('\n2. 创建合同（用于关联发票）...');
  const contractRes = await request('POST', '/api/contracts', {
    code: 'HT2026001',
    name: '测试合同 001',
    status: 'approved'
  });
  console.log(`   状态码: ${contractRes.statusCode}`);
  const contractId = contractRes.body.id;
  console.assert(contractRes.statusCode === 201, '创建合同应该返回 201');

  console.log('\n3. 创建发票 - 增值税专用发票（关联合同）...');
  const invoice1 = await request('POST', '/api/invoices', {
    type: 'special',
    code: '000000000001',
    number: '00000001',
    issueDate: '2026-05-11',
    buyer: { name: '购买方公司', taxId: '911234567890123456' },
    seller: { name: '销售方公司', taxId: '916543210987654321' },
    amount: 100.00,
    taxRate: 0.13,
    contractId: contractId
  });
  console.log(`   状态码: ${invoice1.statusCode}`);
  console.log(`   金额: ${invoice1.body?.amount}, 税额: ${invoice1.body?.taxAmount}, 价税合计: ${invoice1.body?.totalAmount}`);
  const invoice1Id = invoice1.body?.id;
  console.assert(invoice1.statusCode === 201, '创建发票应该返回 201');
  console.assert(invoice1.body?.taxAmount === 13, '100 元 * 13% = 13 元税额');
  console.assert(invoice1.body?.totalAmount === 113, '价税合计应该是 113 元');

  console.log('\n4. 创建发票 - 增值税普通发票（不关联合同）...');
  const invoice2 = await request('POST', '/api/invoices', {
    type: 'normal',
    code: '000000000002',
    number: '00000002',
    issueDate: '2026-05-11',
    buyer: { name: '个人客户', taxId: '123456789012345' },
    seller: { name: '销售方公司', taxId: '916543210987654321' },
    amount: 50.00,
    taxRate: 0.09
  });
  console.log(`   状态码: ${invoice2.statusCode}`);
  console.log(`   金额: ${invoice2.body?.amount}, 税额: ${invoice2.body?.taxAmount}, 价税合计: ${invoice2.body?.totalAmount}`);
  console.assert(invoice2.statusCode === 201, '创建发票应该返回 201');
  console.assert(invoice2.body?.taxAmount === 4.5, '50 元 * 9% = 4.5 元税额');

  console.log('\n5. 创建发票 - 电子发票（1% 税率测试四舍五入）...');
  const invoice3 = await request('POST', '/api/invoices', {
    type: 'electronic',
    code: '000000000003',
    number: '00000003',
    issueDate: '2026-05-11',
    buyer: { name: '电商客户', taxId: '919876543210987654' },
    seller: { name: '销售方公司', taxId: '916543210987654321' },
    amount: 123.45,
    taxRate: 0.01
  });
  console.log(`   状态码: ${invoice3.statusCode}`);
  console.log(`   金额: ${invoice3.body?.amount}, 税额: ${invoice3.body?.taxAmount}, 价税合计: ${invoice3.body?.totalAmount}`);
  console.assert(invoice3.statusCode === 201, '创建发票应该返回 201');

  console.log('\n6. 测试发票代码或号码为空（应该返回 400）...');
  const emptyCode = await request('POST', '/api/invoices', {
    type: 'special',
    code: '',
    number: '00000004',
    issueDate: '2026-05-11',
    buyer: { name: 'Test', taxId: '911234567890123456' },
    seller: { name: 'Seller', taxId: '916543210987654321' },
    amount: 100.00,
    taxRate: 0.13
  });
  console.log(`   状态码: ${emptyCode.statusCode}, 错误: ${emptyCode.body?.error}`);
  console.assert(emptyCode.statusCode === 400, '空代码应该返回 400');

  console.log('\n7. 测试重复发票（应该返回 409）...');
  const duplicate = await request('POST', '/api/invoices', {
    type: 'special',
    code: '000000000001',
    number: '00000001',
    issueDate: '2026-05-11',
    buyer: { name: 'Test', taxId: '911234567890123456' },
    seller: { name: 'Seller', taxId: '916543210987654321' },
    amount: 100.00,
    taxRate: 0.13
  });
  console.log(`   状态码: ${duplicate.statusCode}, 错误: ${duplicate.body?.error}`);
  console.assert(duplicate.statusCode === 409, '重复发票应该返回 409');

  console.log('\n8. 测试无效税率（应该返回 400）...');
  const invalidRate = await request('POST', '/api/invoices', {
    type: 'special',
    code: '000000000004',
    number: '00000004',
    issueDate: '2026-05-11',
    buyer: { name: 'Test', taxId: '911234567890123456' },
    seller: { name: 'Seller', taxId: '916543210987654321' },
    amount: 100.00,
    taxRate: 0.20
  });
  console.log(`   状态码: ${invalidRate.statusCode}, 错误: ${invalidRate.body?.error}`);
  console.assert(invalidRate.statusCode === 400, '无效税率应该返回 400');

  console.log('\n9. 测试不存在的合同（应该返回 404）...');
  const noContract = await request('POST', '/api/invoices', {
    type: 'special',
    code: '000000000005',
    number: '00000005',
    issueDate: '2026-05-11',
    buyer: { name: 'Test', taxId: '911234567890123456' },
    seller: { name: 'Seller', taxId: '916543210987654321' },
    amount: 100.00,
    taxRate: 0.13,
    contractId: 'non-existent-id'
  });
  console.log(`   状态码: ${noContract.statusCode}, 错误: ${noContract.body?.error}`);
  console.assert(noContract.statusCode === 404, '不存在的合同应该返回 404');

  console.log('\n10. 创建未审批的合同并测试...');
  const pendingContract = await request('POST', '/api/contracts', {
    code: 'HT2026002',
    name: '待审批合同',
    status: 'pending'
  });
  const pendingContractId = pendingContract.body?.id;
  
  const pendingContractInvoice = await request('POST', '/api/invoices', {
    type: 'special',
    code: '000000000006',
    number: '00000006',
    issueDate: '2026-05-11',
    buyer: { name: 'Test', taxId: '911234567890123456' },
    seller: { name: 'Seller', taxId: '916543210987654321' },
    amount: 100.00,
    taxRate: 0.13,
    contractId: pendingContractId
  });
  console.log(`   状态码: ${pendingContractInvoice.statusCode}, 错误: ${pendingContractInvoice.body?.error}`);
  console.assert(pendingContractInvoice.statusCode === 400, '未审批合同应该返回 400');

  console.log('\n11. 红冲发票...');
  const redInvoice = await request('POST', '/api/invoices/red', {
    originalInvoiceId: invoice1Id
  });
  console.log(`   状态码: ${redInvoice.statusCode}`);
  console.log(`   金额: ${redInvoice.body?.amount}, 税额: ${redInvoice.body?.taxAmount}, 价税合计: ${redInvoice.body?.totalAmount}`);
  console.log(`   是否红字发票: ${redInvoice.body?.isRedInvoice}`);
  console.assert(redInvoice.statusCode === 201, '红冲应该返回 201');
  console.assert(redInvoice.body?.amount === -100, '红字发票金额应该为负');
  console.assert(redInvoice.body?.isRedInvoice === true, '应该标记为红字发票');

  console.log('\n12. 验证原发票状态已更新为"已红冲"...');
  const updatedOriginal = await request('GET', `/api/invoices/${invoice1Id}`);
  console.log(`   状态码: ${updatedOriginal.statusCode}, 状态: ${updatedOriginal.body?.status}`);
  console.assert(updatedOriginal.body?.status === 'red_invoiced', '原发票状态应该是已红冲');

  console.log('\n13. 测试重复红冲（应该返回 409）...');
  const duplicateRed = await request('POST', '/api/invoices/red', {
    originalInvoiceId: invoice1Id
  });
  console.log(`   状态码: ${duplicateRed.statusCode}, 错误: ${duplicateRed.body?.error}`);
  console.assert(duplicateRed.statusCode === 409, '重复红冲应该返回 409');

  console.log('\n14. 发票查验 - 匹配的情况...');
  const verifyMatch = await request('POST', '/api/invoices/verify', {
    code: '000000000002',
    number: '00000002',
    amount: 50.00,
    taxAmount: 4.50
  });
  console.log(`   状态码: ${verifyMatch.statusCode}`);
  console.log(`   结果: ${verifyMatch.body?.valid}, 消息: ${verifyMatch.body?.message}`);
  console.assert(verifyMatch.statusCode === 200, '匹配应该返回 200');
  console.assert(verifyMatch.body?.valid === true, '应该是有效的');

  console.log('\n15. 发票查验 - 不匹配的情况（应该返回 400）...');
  const verifyMismatch = await request('POST', '/api/invoices/verify', {
    code: '000000000002',
    number: '00000002',
    amount: 51.00,
    taxAmount: 4.50
  });
  console.log(`   状态码: ${verifyMismatch.statusCode}`);
  console.log(`   结果: ${verifyMismatch.body?.valid}, 消息: ${verifyMismatch.body?.message}`);
  console.assert(verifyMismatch.statusCode === 400, '不匹配应该返回 400');
  console.assert(verifyMismatch.body?.valid === false, '应该是无效的');

  console.log('\n16. 查询所有发票...');
  const allInvoices = await request('GET', '/api/invoices');
  console.log(`   状态码: ${allInvoices.statusCode}, 数量: ${allInvoices.body?.length}`);
  console.assert(allInvoices.statusCode === 200, '查询应该返回 200');

  console.log('\n=== 所有测试完成 ===');
}

runTests().catch(console.error);
