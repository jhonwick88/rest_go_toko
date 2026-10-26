@echo off
title Restart Service TokoPintar (restToko)
color 0E
cd /d "%~dp0"

:: Cek Hak Akses Administrator
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo.
    echo =======================================================================
    echo  [PERINGATAN] HAK AKSES ADMINISTRATOR DIBUTUHKAN!
    echo =======================================================================
    echo.
    echo  Harap klik KANAN pada file ini lalu pilih:
    echo  "Run as administrator" (Jalankan sebagai administrator)
    echo.
    echo =======================================================================
    echo.
    pause
    exit /b 1
)

echo =======================================================================
echo              RESTART SERVICE TOKOPINTAR (restToko)
echo =======================================================================
echo.

echo [1/2] Menghentikan service restToko...
net stop restToko >nul 2>&1
timeout /t 2 /nobreak >nul

echo [2/2] Menyalakan kembali service restToko...
net start restToko
if %errorlevel% equ 0 (
    echo.
    echo [SUKSES] Service restToko BERHASIL DIMULAI ULANG (RESTART BERHASIL)!
) else (
    echo.
    echo [GAGAL] Gagal memulai kembali service restToko.
)

echo.
echo =======================================================================
pause
