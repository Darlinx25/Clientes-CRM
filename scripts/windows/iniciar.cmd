@echo off
rem Inicia Historial manualmente (en vez del autoarranque).
rem Todo queda junto al .exe: la base en data\ y las fotos en photos\.
setlocal
cd /d "%~dp0"

rem Opcional: zona horaria de recordatorios (ej. Montevideo).
rem set REMINDER_TIMEZONE=America/Montevideo

rem Opcional: cambiar el puerto (default 7300).
rem set PORT=7300

start "" "%~dp0historial.exe"
endlocal