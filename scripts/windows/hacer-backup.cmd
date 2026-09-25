@echo off
rem Hace un backup (copia de seguridad) de la base de datos y las fotos.
rem Crea una carpeta "backups" junto al .exe con un archivo historial-fecha.zip
rem Puede ejecutarse aunque la app este corriendo (saca una copia consistente).
setlocal
cd /d "%~dp0"

if not exist "%~dp0historial.exe" (
  echo.
  echo  No se encontro historial.exe en esta carpeta.
  pause
  exit /b 1
)

echo.
echo  Creando backup...
echo.

historial.exe backup "%~dp0backups"

if errorlevel 1 (
  echo.
  echo  El backup fallo. Vea el mensaje de arriba.
  pause
  exit /b 1
)

for %%f in ("%~dp0backups\historial-*.zip") do (
  echo  Se creo: %%~nxf
  echo  Carpeta: %%~dpf
)

echo.
echo  Recuerde copiar el archivo a otra carpeta/disco/USB para que el
echo  backup no quede solo en la misma maquina.
echo.
pause
endlocal