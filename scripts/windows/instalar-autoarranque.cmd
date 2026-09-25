@echo off
rem ============================================================
rem  Instala el arranque automatico de Historial al encender la PC
rem  Crea una tarea "Al iniciar el equipo" que ejecuta historial.exe
rem  aunque nadie inicie sesion (la PC es un servidor sin usuarios).
rem  Ejecutar UNA sola vez con permisos de administrador:
rem    clic derecho sobre este archivo -> "Ejecutar como administrador"
rem ============================================================
setlocal
cd /d "%~dp0"

net session >nul 2>&1
if errorlevel 1 (
  echo.
  echo  Necesita permisos de administrador.
  echo  Clic derecho sobre este archivo -^> "Ejecutar como administrador".
  pause
  exit /b 1
)

if not exist "%~dp0historial.exe" (
  echo.
  echo  No se encontro historial.exe en esta carpeta.
  echo  Copie el .exe aqui antes de instalar el autoarranque.
  pause
  exit /b 1
)

echo.
echo  Se creara la tarea "Historial autoarranque".
echo  Windows pedira la contrasena de usuario de esta PC
echo  (la guarda solo Windows, para poder ejecutar el programa
echo   aunque no haya nadie iniciando sesion).
echo.

schtasks /Create /TN "Historial autoarranque" /TR "%~dp0historial.exe" /SC ONSTART /RU "%USERDOMAIN%\%USERNAME%" /RL LIMITED /F

if errorlevel 1 (
  echo.
  echo  No se pudo crear la tarea. Revise los pasos e intente de nuevo.
  pause
  exit /b 1
)

echo.
echo  Listo. Al prender la PC, Historial se iniciara solo.
echo  La web queda en http://localhost:7300 y accesible desde
echo  la red local en http://IP-de-esta-pc:7300
echo.
pause
endlocal