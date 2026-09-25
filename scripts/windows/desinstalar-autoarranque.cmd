@echo off
rem Desinstala la tarea de autoarranque de Historial.
rem Ejecutar como administrador: clic derecho -> "Ejecutar como administrador"
setlocal
net session >nul 2>&1
if errorlevel 1 (
  echo.
  echo  Necesita permisos de administrador.
  pause
  exit /b 1
)

schtasks /Delete /TN "Historial autoarranque" /F

echo.
echo  Autoarranque desinstalado. La app ya no se abrira sola al prender la PC.
pause
endlocal