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

@keyframes bounceIn {
  0% { transform: scale(0); opacity: 0; }
  50% { transform: scale(1.1); opacity: 1; }
  100% { transform: scale(1); }
}

@keyframes float {
  0%, 100% { transform: translateY(0px); }
  50% { transform: translateY(-10px); }
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>

## 🌟 Mengapa BACKEND-GOLANG?
<div style="animation: fadeIn 1s ease-out; text-align: left; max-width: 800px; margin: 0 auto;">
  <p style="font-size: 1.1em; line-height: 1.7; color: #333;">
    Bayangkan sebuah sistem di mana mahasiswa bisa laporkan prestasi dengan mudah, dosen wali verifikasi cepat, dan admin kelola semuanya tanpa ribet. Itulah <strong>BACKEND-GOLANG</strong> – dibuat untuk Universitas Airlangga, tapi scalable untuk kampus mana pun!
  </p>
  <ul style="font-size: 1em; color: #555; line-height: 1.8;">
    <li>🔐 <strong>RBAC Auth</strong>: Role-based access – Admin full power, Mahasiswa CRUD own, Dosen verifikasi only.</li>
    <li>⚡ <strong>JWT Secure</strong>: Token-based login dengan refresh – aman, stateless, expires 24h.</li>
    <li>🗄️ <strong>Hybrid DB</strong>: Postgres relasional (users/roles), Mongo flexible (prestasi dinamis).</li>
    <li>🎯 <strong>Workflow Smart</strong>: Draft → Submitted → Verified/Rejected, dengan notifications & attachments.</li>
    <li>📊 <strong>Analytics Ready</strong>: Statistik prestasi per role, top mahasiswa, & reports mudah.</li>
  </ul>
</div>

## 🛠 Tech Stack
<div style="display: flex; flex-wrap: wrap; justify-content: center; gap: 15px; margin: 30px 0; animation: fadeInUp 1s ease-out 0.8s both;">
  <span class="animated-badge" style="background: linear-gradient(45deg, #00ADD8, #007acc); color: white; padding: 10px 20px; border-radius: 25px; font-weight: bold; box-shadow: 0 4px 15px rgba(0,173,216,0.3);">
    <img src="https://img.shields.io/badge/Go-00ADD8?style=social&logo=go&logoColor=white" alt="Go" width="20" height="20" style="margin-right: 5px;"> Go (Fiber)
  </span>
  <span class="animated-badge" style="background: linear-gradient(45deg, #316192, #3366cc); color: white; padding: 10px 20px; border-radius: 25px; font-weight: bold; box-shadow: 0 4px 15px rgba(49,97,146,0.3);">
    <img src="https://img.shields.io/badge/PostgreSQL-316192?style=social&logo=postgresql&logoColor=white" alt="Postgres" width="20" height="20" style="margin-right: 5px;"> PostgreSQL
  </span>
  <span class="animated-badge" style="background: linear-gradient(45deg, #4EA94B, #66bb6a); color: white; padding: 10px 20px; border-radius: 25px; font-weight: bold; box-shadow: 0 4px 15px rgba(78,169,75,0.3);">
    <img src="https://img.shields.io/badge/MongoDB-4EA94B?style=social&logo=mongodb&logoColor=white" alt="Mongo" width="20" height="20" style="margin-right: 5px;"> MongoDB
  </span>
  <span class="animated-badge" style="background: linear-gradient(45deg, #000, #333); color: white; padding: 10px 20px; border-radius: 25px; font-weight: bold; box-shadow: 0 4px 15px rgba(0,0,0,0.3);">
    <img src="https://img.shields.io/badge/JWT-000000?style=social&logo=json-web-tokens&logoColor=white" alt="JWT" width="20" height="20" style="margin-right: 5px;"> JWT Auth
  </span>
</div>

<style>
.animated-badge {
  animation: slideInLeft 0.8s ease-out;
}

@keyframes slideInLeft {
  from { transform: translateX(-100%); opacity: 0; }
  to { transform: translateX(0); opacity: 1; }
}
</style>

## 🚀 Quick Start
<div style="animation: fadeInUp 1s ease-out 1s both; text-align: center; max-width: 800px; margin: 0 auto;">
  <h3 style="color: #FF6B35; font-size: 1.5em;">Mulai dalam 5 Menit!</h3>
  <div style="background: #f8f9fa; padding: 20px; border-radius: 10px; border-left: 5px solid #00ADD8; margin: 20px 0;">
    <pre style="background: none; border: none; font-family: monospace; color: #333; text-align: left; margin: 0;">