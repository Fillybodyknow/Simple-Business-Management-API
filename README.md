# Simple Business Management API

ระบบจัดการธุรกิจขนาดเล็ก (สินค้าลูกค้าออเดอร์) พัฒนาโดยใช้ **Go (Gin)** และ **MongoDB**  
รองรับ Authentication (JWT), Role (Admin, Staff, Customer) และระบบ CRUD ครบถ้วน

---

## 📦 Features

- ✅ Authentication + JWT
- ✅ Role-based access (Admin / Staff / Customer)
- ✅ CRUD: Product / Order / Customer
- ✅ สร้างเลข Tracking Number อัตโนมัติ
- ✅ ระบบ stock อัปเดตเมื่อมีการสั่งซื้อ
- ✅ Staff เห็นเฉพาะออเดอร์ของตนเอง

---

## 🛠 Tech Stack

| Layer         | Tools                     |
|---------------|----------------------------|
| Language      | Golang (Go 1.21+)         |
| Web Framework | [Gin](https://github.com/gin-gonic/gin) |
| Database      | MongoDB (NoSQL)           |
| Auth          | JWT (JSON Web Token)      |
| Env           | godotenv                  |

---

## ⚙️ Installation

```bash
# 1. Clone the project
git clone https://github.com/Fillybodyknow/simple-business-management-api.git
cd simple-business-management-api

# 2. Create .env file
cp .env.example .env
