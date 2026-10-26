@echo off
title Mematikan Service TokoPintar (restToko)
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
echo              MEMATIKAN SERVICE TOKOPINTAR (restToko)
echo =======================================================================
echo.

sc query restToko | findstr /i "STATE" | findstr /i "STOPPED" >nul 2>&1
if %errorlevel% equ 0 (
    echo [INFO] Service restToko SUDAH MATI / NONAKTIF (STOPPED).
) else (
    echo Menghentikan service restToko...
    net stop restToko
    if %errorlevel% equ 0 (
        echo.
        echo [SUKSES] Service restToko BERHASIL DIMATIKAN!
    ) else (
        echo.
        echo [GAGAL] Tidak dapat mematikan service restToko.
    )
)

echo.
echo =======================================================================
pause
