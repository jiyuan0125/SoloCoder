export async function sendAppointmentConfirmation(clientPhone: string, details: {
  counselorName: string;
  date: string;
  startTime: string;
}): Promise<boolean> {
  console.log(`[SMS] 发送通知给 ${clientPhone}:`);
  console.log(`       您的预约已确认：${details.counselorName} 咨询师，${details.date} ${details.startTime}`);
  
  return true;
}
