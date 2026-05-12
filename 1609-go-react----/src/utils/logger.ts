export function logPushNotification(messageId: string, recipientId: string, subject: string): void {
  console.log(`[PUSH] ${new Date().toISOString()} - Message ID: ${messageId}, Recipient: ${recipientId}, Subject: ${subject}`);
}

export function logEmailNotification(messageId: string, recipientId: string, subject: string): void {
  console.log(`[EMAIL] ${new Date().toISOString()} - Message ID: ${messageId}, Recipient: ${recipientId}, Subject: ${subject}`);
}
