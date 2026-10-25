@echo off
echo ========================================================
echo UPDATE SERVICE TOKOPINTAR POS BACKEND (REST_GO_TOKO)
echo ========================================================
echo.

:: Check for admin rights
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo [ERROR] Harap klik kanan file ini dan pilih 'Run as administrator' (Jalankan sebagai administrator)!
    echo.
    pause
    exit /b 1
)

echo [1/4] Menghentikan service restToko...
net stop restToko
timeout /t 2 /nobreak >nul

echo [2/4] Mengompilasi backend terbaru...
cd /d "H:\FlutterProject\rest_go_toko"
go build -ldflags="-s -w" -o "restToko.exe" .

echo [3/4] Menyalin binary dan konfigurasi .env terbaru ke folder service...
copy /Y "H:\FlutterProject\rest_go_toko\restToko.exe" "C:\Program Files\TokoPintarPOS\service\restToko.exe"
copy /Y "H:\FlutterProject\rest_go_toko\.env" "C:\Program Files\TokoPintarPOS\service\.env"

echo [4/4] Menjalankan kembali service restToko...
net start restToko

echo. 
echo ========================================================
echo SUKSES! Service restToko berhasil diperbarui dan berjalan!
echo ========================================================
pause
