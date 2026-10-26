@echo off
title Restart Service Bell Pintar (bellServer)
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
echo             RESTART SERVICE BELL PINTAR (bellServer)
echo =======================================================================
echo.

echo [1/2] Menghentikan service bellServer...
net stop bellServer >nul 2>&1
timeout /t 2 /nobreak >nul

echo [2/2] Menyalakan kembali service bellServer...
net start bellServer
if %errorlevel% equ 0 (
    echo.
    echo [SUKSES] Service bellServer BERHASIL DIMULAI ULANG (RESTART BERHASIL)!
) else (
    echo.
    echo [GAGAL] Gagal memulai kembali service bellServer.
)

echo.
echo =======================================================================
pause
