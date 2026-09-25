@echo off
rem Abre el puerto 7300 en el Firewall de Windows para que otras
rem computadoras/telofonos de la red puedan entrar a Historial.
rem Ejecutar UNA sola vez con administrador:
rem   clic derecho -> "Ejecutar como administrador"
setlocal
net session >nul 2>&1
if errorlevel 1 (
  echo.
  echo  Necesita permisos de administrador.
  echo  Clic derecho sobre este archivo -^> "Ejecutar como administrador".
  pause
  exit /b 1
)

netsh advfirewall firewall add rule name="Historial 7300" dir=in action=allow protocol=TCP localport=7300

if errorlevel 1 (
  echo.
  echo  No se pudo agregar la regla de firewall.
  pause
  exit /b 1
)

echo.
echo  Regla agregada. Desde otras PCs de la red ya se puede entrar a:
echo    http://IP-de-esta-pc:7300
echo  (ver la IP con el comando "ipconfig" en esta maquina).
echo.
pause
endlocal