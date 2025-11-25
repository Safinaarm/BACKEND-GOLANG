# BACKEND-GOLANG 🚀

<div align="center">
  <h1 style="background: linear-gradient(45deg, #00ADD8, #FF6B35, #4EA94B); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; animation: glow 2s ease-in-out infinite alternate, slideInDown 1s ease-out;">
    Backend Golang: Sistem Pelaporan Prestasi Mahasiswa
  </h1>
  
  <p style="font-size: 1.3em; color: #666; animation: fadeInUp 1s ease-out 0.3s both; max-width: 800px; margin: 0 auto; line-height: 1.6;">
    Aplikasi backend modern berbasis <strong>Go</strong> untuk mengelola prestasi mahasiswa dengan <strong>RBAC autentikasi</strong>, <strong>JWT security</strong>, dan workflow verifikasi yang seamless. Dibangun dengan <strong>PostgreSQL</strong> untuk data struktural & <strong>MongoDB</strong> untuk prestasi dinamis – cepat, scalable, dan siap produksi!
  </p>

  <div style="animation: pulse 2s infinite; margin: 20px 0;">
    <img src="https://img.shields.io/badge/⭐%20Star%20Us-FF6B35?style=for-the-badge&logo=github&logoColor=white" alt="Star">
    <img src="https://img.shields.io/badge/%F0%9F%94%A5%20Fork-00ADD8?style=for-the-badge&logo=github&logoColor=white" alt="Fork">
    <img src="https://img.shields.io/badge/%F0%9F%9A%80%20Live-4EA94B?style=for-the-badge&logo=go&logoColor=white" alt="Live">
  </div>

  <div style="animation: bounceIn 1s ease-out 0.6s both;">
    <img src="https://github-readme-stats.vercel.app/api?username=Safinaarm&show_icons=true&theme=radical&hide_border=true&bg_color=0d1117&title_color=00ADD8&text_color=9eb2b8&hide=stars,prs" width="420" height="240">
  </div>
</div>

---

## 📌 Deskripsi Singkat
Backend ini dibangun menggunakan **Golang**, sebagai bagian dari **UAS Backend Development**.  
Sistem memanfaatkan dua database secara bersamaan:

- **PostgreSQL** — menyimpan data relasional (user, role, autentikasi).
- **MongoDB** — menyimpan data prestasi yang fleksibel dan dinamis.

Struktur dibuat agar scalable untuk kebutuhan kampus atau organisasi lainnya.

---

## 🌟 Mengapa BACKEND-GOLANG?
<div style="animation: fadeIn 1s ease-out; text-align: left; max-width: 800px; margin: 0 auto;">
  <p style="font-size: 1.1em; line-height: 1.7; color: #333;">
    Sistem dirancang untuk mempermudah mahasiswa melaporkan prestasi, dosen melakukan verifikasi, dan admin mengelola keseluruhan proses dengan efisien.
  </p>
  <ul style="font-size: 1em; color: #555; line-height: 1.8;">
    <li>🔐 <strong>RBAC Auth</strong>: Role-based access – Admin, Mahasiswa, dan Dosen memiliki izin berbeda.</li>
    <li>⚡ <strong>JWT Secure</strong>: Token login + refresh token yang aman dan stateless.</li>
    <li>🗄️ <strong>Hybrid DB</strong>: Kombinasi PostgreSQL dan MongoDB.</li>
    <li>🎯 <strong>Smart Workflow</strong>: Draft → Submitted → Verified/Rejected.</li>
    <li>📊 <strong>Analytics Ready</strong>: Statistik dan laporan siap dikembangkan.</li>
  </ul>
</div>

---

## 🛠 Tech Stack

<div style="display: flex; flex-wrap: wrap; justify-content: center; gap: 15px; margin: 30px 0;">
  <span class="animated-badge" style="background: linear-gradient(45deg, #00ADD8, #007acc); color: white; padding: 10px 20px; border-radius: 25px; font-weight: bold;">
    Go (Fiber)
  </span>
  <span class="animated-badge" style="background: linear-gradient(45deg, #316192, #3366cc); color: white; padding: 10px 20px; border-radius: 25px; font-weight: bold;">
    PostgreSQL
  </span>
  <span class="animated-badge" style="background: linear-gradient(45deg, #4EA94B, #66bb6a); color: white; padding: 10px 20px; border-radius: 25px; font-weight: bold;">
    MongoDB
  </span>
  <span class="animated-badge" style="background: linear-gradient(45deg, #000, #333); color: white; padding: 10px 20px; border-radius: 25px; font-weight: bold;">
    JWT Auth
  </span>
</div>

---

## 🚀 Quick Start

