# BACKEND-GOLANG 🚀

<div align="center">
  <h1 style="background: linear-gradient(45deg, #00ADD8, #FF6B35, #4EA94B); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; animation: glow 2s ease-in-out infinite alternate, slideInDown 1s ease-out;">
    Backend Golang: Sistem Pelaporan Prestasi Mahasiswa
  </h1>
  
  <p style="font-size: 1.2em; color: #666; animation: fadeInUp 1s ease-out 0.3s both; max-width: 700px; margin: 0 auto; line-height: 1.6;">
    Backend modern berbasis Go dengan RBAC + JWT untuk login role-based (Admin, Mahasiswa, Dosen Wali). Gunakan PostgreSQL untuk users/roles & MongoDB untuk prestasi dinamis. Cepat, aman, scalable!
  </p>

  <div style="animation: pulse 2s infinite; margin: 20px 0;">
    <img src="https://img.shields.io/badge/⭐%20Star-FF6B35?style=for-the-badge&logo=github&logoColor=white" alt="Star">
    <img src="https://img.shields.io/badge/%F0%9F%94%A5%20Fork-00ADD8?style=for-the-badge&logo=github&logoColor=white" alt="Fork">
    <img src="https://img.shields.io/badge/%F0%9F%9A%80%20Live-4EA94B?style=for-the-badge&logo=go&logoColor=white" alt="Live">
  </div>
</div>

<style>
@keyframes glow {
  from { text-shadow: 0 0 20px #00ADD8; }
  to { text-shadow: 0 0 30px #FF6B35; }
}

@keyframes slideInDown {
  from { transform: translateY(-50px); opacity: 0; }
  to { transform: translateY(0); opacity: 1; }
}

@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes pulse {
  0% { transform: scale(1); }
  50% { transform: scale(1.05); }
  100% { transform: scale(1); }
}
</style>

## 🛠 Tech Stack
<div style="display: flex; flex-wrap: wrap; justify-content: center; gap: 15px; margin: 30px 0; animation: fadeInUp 1s ease-out 0.6s both;">
  <span style="background: linear-gradient(45deg, #00ADD8, #007acc); color: white; padding: 10px 15px; border-radius: 20px; font-weight: bold; box-shadow: 0 4px 15px rgba(0,173,216,0.3); animation: slideInLeft 0.8s ease-out;">
    Go (Fiber)
  </span>
  <span style="background: linear-gradient(45deg, #316192, #3366cc); color: white; padding: 10px 15px; border-radius: 20px; font-weight: bold; box-shadow: 0 4px 15px rgba(49,97,146,0.3); animation: slideInLeft 0.8s ease-out 0.2s;">
    PostgreSQL
  </span>
  <span style="background: linear-gradient(45deg, #4EA94B, #66bb6a); color: white; padding: 10px 15px; border-radius: 20px; font-weight: bold; box-shadow: 0 4px 15px rgba(78,169,75,0.3); animation: slideInLeft 0.8s ease-out 0.4s;">
    MongoDB
  </span>
  <span style="background: linear-gradient(45deg, #000, #333); color: white; padding: 10px 15px; border-radius: 20px; font-weight: bold; box-shadow: 0 4px 15px rgba(0,0,0,0.3); animation: slideInLeft 0.8s ease-out 0.6s;">
    JWT Auth
  </span>
</div>

<style>
@keyframes slideInLeft {
  from { transform: translateX(-100%); opacity: 0; }
  to { transform: translateX(0); opacity: 1; }
}
</style>

## 🚀 Quick Start
<div style="animation: fadeInUp 1s ease-out 0.8s both; text-align: center; max-width: 700px; margin: 0 auto;">
  <h3 style="color: #FF6B35; font-size: 1.4em;">Mulai dalam 5 Menit!</h3>
  <div style="background: #f8f9fa; padding: 15px; border-radius: 10px; border-left: 5px solid #00ADD8; margin: 20px 0; font-family: monospace; color: #333; text-align: left;">
<pre style="margin: 0; overflow-x: auto;">
git clone https://github.com/Safinaarm/BACKEND-GOLANG.git
cd BACKEND-GOLANG
cp .env.example .env  # Edit DB creds
go mod tidy
go run main.go
</pre>
  </div>
  <p style="color: #666;">Server live di <code style="background: #00ADD8; color: white; padding: 3px 6px; border-radius: 4px;">localhost:3000</code> – Test dengan Postman!</p>
</div>

<div align="center" style="margin: 40px 0; animation: fadeInUp 1s ease-out 1s both;">
  <a href="https://github.com/Safinaarm/BACKEND-GOLANG/issues" style="background: linear-gradient(45deg, #667eea, #764ba2); color: white; padding: 12px 30px; border-radius: 30px; text-decoration: none; font-weight: bold; box-shadow: 0 4px 15px rgba(102,126,234,0.4); transition: all 0.3s ease;">
    🚀 Star & Contribute!
  </a>
  <p style="color: #4EA94B; font-weight: bold; font-size: 1.1em; margin-top: 20px; animation: glow 2s ease-in-out infinite alternate;">
    Oleh: <strong>Safina Rahmani Maulidiyah</strong> | NIM: <strong>434231034</strong>
  </p>
  <p style="color: #888; margin-top: 10px; font-size: 0.9em;">
    Dibuat untuk UAS Backend Lanjut, Universitas Airlangga | Open Source 2025
  </p>
</div>

<style>
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>