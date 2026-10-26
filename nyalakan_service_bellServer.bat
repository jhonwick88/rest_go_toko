@echo off
title Menyalakan Service Bell Pintar (bellServer)
color 0A
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
echo             MENYALAKAN SERVICE BELL PINTAR (bellServer)
echo =======================================================================
echo.

sc query bellServer | findstr /i "STATE" | findstr /i "RUNNING" >nul 2>&1
if %errorlevel% equ 0 (
    echo [INFO] Service bellServer SUDAH BERJALAN (RUNNING).
) else (
    echo Memulai service bellServer...
    net start bellServer
    if %errorlevel% equ 0 (
        echo.
        echo [SUKSES] Service bellServer BERHASIL DINYALAKAN!
    ) else (
        echo.
        echo [GAGAL] Tidak dapat menyalakan service bellServer.
        echo Pastikan service telah terdaftar di services.msc.
    )
)

echo.
echo =======================================================================
pause
