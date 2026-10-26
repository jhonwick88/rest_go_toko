@echo off
title Mematikan Service Bell Pintar (bellServer)
color 0C
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
echo             MEMATIKAN SERVICE BELL PINTAR (bellServer)
echo =======================================================================
echo.

sc query bellServer | findstr /i "STATE" | findstr /i "STOPPED" >nul 2>&1
if %errorlevel% equ 0 (
    echo [INFO] Service bellServer SUDAH MATI / NONAKTIF (STOPPED).
) else (
    echo Menghentikan service bellServer...
    net stop bellServer
    if %errorlevel% equ 0 (
        echo.
        echo [SUKSES] Service bellServer BERHASIL DIMATIKAN!
    ) else (
        echo.
        echo [GAGAL] Tidak dapat mematikan service bellServer.
    )
)

echo.
echo =======================================================================
pause
