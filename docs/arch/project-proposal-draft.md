# **สรุปโปรเจกต์ที่จะเปลี่ยนสำหรับ Term Project (Software Architecture)**

## **1\) ระบบเกี่ยวกับอะไร**

**Eventory** เป็นแพลตฟอร์มจัดการ tournament ที่รองรับการระดมทุนจาก sponsor

ผู้ใช้ล็อกอินด้วยบัญชีเดียว แล้วสลับโหมดการใช้งานได้ 3 แบบ:

- **Competitor**: เข้าร่วม tournament / รวมทีม  
- **Organizer**: สร้างและดำเนิน tournament  
- **Sponsor**: สนับสนุนเงินทุน / ซื้อแพ็กเกจ sponsor

## **2\) Business Use Cases หลัก (3 ข้อ)**

| ID | Use Case | Actor | ผลลัพธ์โดยย่อ |
| :---- | :---- | :---- | :---- |
| UC-01 | Create a Tournament | Organizer | สร้าง tournament *(โดยเลือกได้ว่าจะทำ crowdfunding ก่อน หรือเปิดรับสมัครผู้เข้าแข่งขันเลยก็ได้) \+*  สร้างฟอร์มสมัคร tournament |
| UC-02 | Sponsor a Tournament | Sponsor | เลือกแพ็กเกจ/pledge \+ ชำระเงิน \+ แสดง sponsor บนหน้า tournament / อัปเดตความคืบหน้า funding |
| UC-03 | Submit Registration | Team Captain (Competitor) | ล็อกสมาชิกทีม (ในกรณีแข่งแบบทีม) \+ ชำระค่าสมัคร (ถ้ามี) แล้วส่งฟอร์มสมัครให้ organizer |
| UC-04 | Record Match Results | Organizer (หรือ tournament staff) | อัพเดต match result แบบราย match และแสดงผลลัพธ์ในหน้า tournament |

## **3\) องค์ประกอบหลัก (แผน microservice คร่าว ๆ)**

**Services**

- **Account**: identity, profiles, access control  
- **Tournament**: tournament, team/lobby, registration, match/bracket  
- **Payment**: payment, refund, ledger, callback จาก payment gateway  
- **Notification**: ส่งการแจ้งเตือน (นอกแอปใช้ Email, ในแอปใช้ SSE)

**External Services**

- Authentication (ใช้ Supabase Auth)  
- Payment Gateway  
- Object Storage (S3-compatible, สำหรับโลโก้ sponsor)

**Service Communication**

- Client \-\> API Gateway \-\> REST (บริการที่เปิดให้ UI)  
- Internal synchronous call เช่น Tournament \-\> Payment ผ่าน **gRPC**  
- Internal asynchronous call เช่น หลังชำระเงินสำเร็จ / หลังบันทึกผลแมตช์ มีการส่ง Notification ผ่าน **Message Broker** *(ให้ Notification Service consume queue และทำระบบ retry & outbox เพื่อ delivery ที่ eventually consistent)*

**Databases**

- By default ข้อมูลส่วนใหญ่อยากได้ความเป็น transaction ต้องการ ACID property จะใช้ **PostgreSQL**  
- ข้อมูลที่มาจาก custom registration form (ฟอร์มสมัครที่มีโครงสร้างยืดหยุ่นตามแต่ละ tournament) จะใช้ **MongoDB**

**API Gateway \+ Service Discovery**

- ใช้ **Traefik** เป็น API Gateway ที่ discover route จาก labels แบบ dynamic  
- ใช้ **Docker Swarm** สำหรับ production environment เพราะดูแลง่ายกว่า Kubernetes สำหรับทีมที่เล็กและระยะเวลาที่จำกัด พร้อม service discovery ผ่าน Swarm DNS (service เรียกกันด้วยชื่อ) โดยไม่ต้องรัน registry แยก
