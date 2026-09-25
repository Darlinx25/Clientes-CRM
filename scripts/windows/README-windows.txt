=====================================================================
 HISTORIAL - Instalacion en una PC con Windows (servidor)
 Usuario, contrasena, base y fotos => todo se guarda en esta carpeta.
 Sistema: un solo archivo historial.exe. No instala nada.
=====================================================================

1) PRIMERA VEZ
   - Copie la carpeta completa a un disco local, por ejemplo:
       C:\Historial\
     (NO la deje en OneDrive/Dropbox/unidad de red: la base no debe
     vivir en carpetas sincronizadas).
   - Doble clic sobre historial.exe.
   - En el navegador abra:  http://localhost:7300
   - Entrar con el usuario admin (primera vez: admin / admin).
     IMPORTANTE: cambie esa contrasena cuanto antes
     (Configuracion -> Cambiar Contrasena).

2) AUTOARRANQUE AL PRENDER LA PC (aunque nadie inicie sesion)
   - Clic derecho sobre instalar-autoarranque.cmd -> "Ejecutar como
     administrador" y siga los pasos (pide la contrasena de Windows
     una sola vez, la guarda solo Windows).
   - Para quitarlo: desinstalar-autoarranque.cmd

3) ACCESO DESDE OTRAS COMPUTADORAS / TELEFONOS DE LA RED
   - Clic derecho sobre abrir-firewall.cmd -> "Ejecutar como
     administrador" (habilita el puerto 7300).
   - Averigue la IP de esta PC con "ipconfig" (IPv4).
   - Desde las otras computadoras entre a:
       http://IP-de-esta-pc:7300

4) BACKUP (copia de seguridad)
   - La forma recomendada es desde la web, como administrador:
       Configuracion -> "Hacer backup"
     Genera en C:\Historial\backups\ un archivo
       historial-FECHA.zip (base de datos + fotos)
     Se puede hacer con la app corriendo; es consistente.
   - Alternativa por consola: doble clic en hacer-backup.cmd
   - IMPORTANTE: copie el .zip a OTRA maquina/USB/OneDrive. Un backup
     guardado solo en el mismo disco no lo protege de un fallo del disco.

5) ACTUALIZAR LA APP
   - Entre a la web y, de ser posible, haga un backup (paso 4).
   - Cierre la app (si esta corriendo):
       Administrador de tareas -> finalizar "Historial".
     O: desde Instalador de tareas, "deshabilitar" la tarea
       "Historial autoarranque", cerrar la app y volver a habilitarla.
   - Reemplace historial.exe por la version nueva (copie encima).
     NO borre las carpetas data\ ni photos\: ahi viven los datos.
   - Vuelva a abrir historial.exe (doble clic). Las actualizaciones de
     la base se aplican solas en el primer arranque.
   - Compruebe que la web abre bien (http://localhost:7300).

6) DATOS
   - data\meerkat.db  -> la base de datos (NO copiar a mano con la app
                         corriendo; use el boton de backup).
   - photos\          -> las fotos de clientes.
   - backups\         -> los backups generados.

7) NOTAS
   - Windows puede mostrar un aviso la primera vez al abrir historial.exe
     ("Windows protegi o su equipo"): "Mas informacion" -> "Ejecutar de
     todas formas". Es normal, el archivo no esta firmado digitalmente.
   - Puertos/entorno opcionales (variables del sistema):
       PORT=7300, REMINDER_TIMEZONE=America/Montevideo,
       BACKUP_DIR=C:\Historial\backups
   - La app es liviana (un proceso) y aguanta varios usuarios a la vez.
=====================================================================