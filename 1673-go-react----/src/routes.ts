import { Router } from 'express';
import * as controller from './controllers/userController';

const router = Router();

router.post('/users', controller.createUser);

router.get('/users/:userId/accounts', controller.listUserAccounts);
router.post('/users/:userId/accounts', controller.createSubAccount);

router.get('/accounts/:accountId', controller.getAccount);
router.post('/accounts/:accountId/recharge', controller.recharge);
router.get('/accounts/:accountId/transactions', controller.listAccountTransactions);
router.get('/accounts/:accountId/verify', controller.verifyBalance);

router.post('/transfers', controller.transferAccounts);

router.post('/users/:userId/withdraw', controller.createWithdraw);

router.post('/escrow/pay', controller.escrowPay);
router.post('/escrow/:escrowId/ship', controller.escrowConfirmShipment);
router.post('/escrow/:escrowId/receive', controller.escrowConfirmReceipt);
router.post('/escrow/:escrowId/refund', controller.escrowRefund);
router.get('/escrow/:escrowId', controller.getEscrowTransaction);

export default router;
