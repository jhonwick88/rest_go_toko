@echo off
setlocal
title Kelola Service TokoPintar POS (restToko)
color 0F
cd /d "%~dp0"

:: ========================================================
:: 1. Cek Hak Akses Administrator (Windows 10/11)
:: ========================================================
net session >nul 2>&1
if %errorlevel% neq 0 goto :NOT_ADMIN
goto :MENU

:NOT_ADMIN
cls
echo =======================================================================
echo              [PERINGATAN] HAK AKSES ADMINISTRATOR DIBUTUHKAN
echo =======================================================================
echo.
echo  Script ini memerlukan izin Administrator untuk mengontrol Service Windows.
echo.
echo  Langkah Mudah:
echo  1. Tutup jendela ini.
echo  2. Klik KANAN pada file: %~nx0
echo  3. Pilih "Run as administrator" [Jalankan sebagai administrator]
echo.
echo =======================================================================
echo.
pause
exit /b 1

:MENU
cls
echo =======================================================================
echo              PENGELOLA SERVICE TOKOPINTAR (restToko)
echo                        Windows 10 / 11
echo =======================================================================
echo.

:: Cek Status Terkini
set "SVC_STATUS=[ TIDAK TERDAFTAR / BELUM DIINSTAL ]"
sc query restToko | findstr /i "STATE" | findstr /i "RUNNING" >nul 2>&1
if %errorlevel% equ 0 (
    set "SVC_STATUS=[ AKTIF / BERJALAN ]"
    goto :SHOW_MENU
)
sc query restToko | findstr /i "STATE" | findstr /i "STOPPED" >nul 2>&1
if %errorlevel% equ 0 (
    set "SVC_STATUS=[ NONAKTIF / MATI ]"
    goto :SHOW_MENU
)

:SHOW_MENU
echo  Status Service Saat Ini : %SVC_STATUS%
echo.
echo -----------------------------------------------------------------------
echo  PILIHAN MENU:
echo -----------------------------------------------------------------------
echo  [1] Nyalakan Service  (Start)
echo  [2] Matikan Service   (Stop)
echo  [3] Restart Service   (Mulai Ulang)
echo  [4] Cek Status Detail (Informasi Lengkap)
echo  [5] Buka Jendela Services.msc
echo  [0] Keluar
echo -----------------------------------------------------------------------
echo.
set "PILIHAN="
set /p PILIHAN="Masukkan nomor pilihan Anda [0-5]: "

if "%PILIHAN%"=="1" goto START_SVC
if "%PILIHAN%"=="2" goto STOP_SVC
if "%PILIHAN%"=="3" goto RESTART_SVC
if "%PILIHAN%"=="4" goto STATUS_SVC
if "%PILIHAN%"=="5" goto OPEN_SERVICES
if "%PILIHAN%"=="0" goto KELUAR

echo.
echo [!] Pilihan tidak valid, silakan masukkan angka 0 sampai 5.
timeout /t 2 >nul
goto MENU

:START_SVC
cls
echo =======================================================================
echo  MENYALAKAN SERVICE restToko...
echo =======================================================================
echo.
sc query restToko | findstr /i "STATE" | findstr /i "RUNNING" >nul 2>&1
if %errorlevel% equ 0 goto :START_ALREADY_RUNNING

net start restToko
if %errorlevel% equ 0 goto :START_SUCCESS
goto :START_FAILED

:START_ALREADY_RUNNING
echo [INFO] Service restToko sudah dalam keadaan BERJALAN.
goto :AFTER_START

:START_SUCCESS
echo.
echo [SUKSES] Service restToko berhasil dinyalakan!
goto :AFTER_START

:START_FAILED
echo.
echo [GAGAL] Gagal menyalakan service restToko.
goto :AFTER_START

:AFTER_START
echo.
pause
goto MENU

:STOP_SVC
cls
echo =======================================================================
echo  MEMATIKAN SERVICE restToko...
echo =======================================================================
echo.
sc query restToko | findstr /i "STATE" | findstr /i "STOPPED" >nul 2>&1
if %errorlevel% equ 0 goto :STOP_ALREADY_STOPPED

net stop restToko
if %errorlevel% equ 0 goto :STOP_SUCCESS
goto :STOP_FAILED

:STOP_ALREADY_STOPPED
echo [INFO] Service restToko sudah dalam keadaan MATI.
goto :AFTER_STOP

:STOP_SUCCESS
echo.
echo [SUKSES] Service restToko berhasil dimatikan!
goto :AFTER_STOP

:STOP_FAILED
echo.
echo [GAGAL] Gagal mematikan service restToko.
goto :AFTER_STOP

:AFTER_STOP
echo.
pause
goto MENU

:RESTART_SVC
cls
echo =======================================================================
echo  RESTART (MULAI ULANG) SERVICE restToko...
echo =======================================================================
echo.
echo [1/2] Menghentikan service restToko...
net stop restToko >nul 2>&1
timeout /t 2 /nobreak >nul

echo [2/2] Menyalakan kembali service restToko...
net start restToko
if %errorlevel% equ 0 goto :RESTART_SUCCESS
goto :RESTART_FAILED

:RESTART_SUCCESS
echo.
echo [SUKSES] Service restToko berhasil dimulai ulang!
goto :AFTER_RESTART

:RESTART_FAILED
echo.
echo [GAGAL] Gagal menyalakan kembali service restToko.
goto :AFTER_RESTART

:AFTER_RESTART
echo.
pause
goto MENU

:STATUS_SVC
cls
echo =======================================================================
echo  DETAIL STATUS SERVICE restToko:
echo =======================================================================
echo.
sc query restToko
echo.
echo -----------------------------------------------------------------------
pause
goto MENU

:OPEN_SERVICES
echo Membuka services.msc...
start services.msc
goto MENU

:KELUAR
cls
echo Terima kasih telah menggunakan layanan TokoPintar.
timeout /t 2 >nul
exit /b 0
