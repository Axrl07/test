# Manual Técnico para la Simulación de Sistemas de Archivos EXT2 y EXT3

## Arquitectura del Sistema
El sistema adopta una arquitectura cliente-servidor, donde el backend es responsable de todas las operaciones del sistema de archivos y expone una API RESTful para la comunicación con el frontend. El backend, desarrollado en Go, se ejecuta en una instancia EC2 de AWS con un sistema operativo basado en Linux (se recomienda Ubuntu). El frontend, alojado en un bucket S3, no se detalla en este documento según las instrucciones proporcionadas.

### Descripción del Backend
- **Lenguaje**: Go (Golang).
- **Finalidad**: Gestionar los sistemas de archivos EXT2 y EXT3, procesar comandos y mantener el estado de discos y particiones virtuales.
- **Despliegue**: Instancia EC2 en AWS con sistema operativo Linux.
- **API**: API RESTful que procesa solicitudes del frontend, como la creación de discos, gestión de particiones, operaciones de archivos y recuperación del sistema.
- **Simulación del Sistema de Archivos**: Utiliza archivos binarios con extensión `.mia` para simular discos, almacenando estructuras como el Master Boot Record (MBR), Extended Boot Record (EBR), superbloques, inodos y bloques.

### Despliegue en AWS
- **Instancia EC2**: Aloja el backend desarrollado en Go, ejecutándose en un sistema operativo Linux. La instancia procesa las solicitudes de la API y realiza las operaciones del sistema de archivos.
- **Bucket S3**: Almacena la aplicación frontend, que se comunica con el backend mediante solicitudes HTTP.
- **Comunicación**: El frontend envía comandos al backend, que los procesa y devuelve respuestas, como resultados de comandos o datos para la visualización del sistema de archivos.

### Diagrama de Arquitectura
El backend interactúa con los archivos `.mia` para simular operaciones de disco, organizando las estructuras en un orden predefinido (por ejemplo, el MBR al inicio, seguido de las particiones). Los endpoints de la API están mapeados a comandos como `MKDISK`, `FDISK` y `MKFS`, garantizando modularidad y escalabilidad.

## Estructuras de Datos
El sistema se basa en un conjunto de estructuras de datos esenciales para simular los sistemas de archivos EXT2 y EXT3, almacenadas en archivos binarios con extensión `.mia`. A continuación, se describen estas estructuras, sus funciones y su organización.

### Master Boot Record (MBR)
El MBR se escribe al inicio de cada archivo `.mia` y contiene metadatos sobre el disco y sus particiones.

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `mbr_tamano` | int | Tamaño total del disco en bytes |
| `mbr_fecha_creacion` | time | Fecha y hora de creación del disco |
| `mbr_dsk_signature` | int | Identificador único del disco (generado aleatoriamente) |
| `dsk_fit` | char | Tipo de ajuste para particiones (B: Best, F: First, W: Worst) |
| `mbr_partitions` | partition[4] | Arreglo de 4 estructuras de partición |

### Partición
Cada estructura de partición dentro del MBR describe una partición primaria o extendida.

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `part_status` | char | Estado de la partición (activa o inactiva) |
| `part_type` | char | Tipo de partición (P: Primaria, E: Extendida) |
| `part_fit` | char | Tipo de ajuste (B, F, W) |
| `part_start` | int | Byte inicial de la partición |
| `part_s` | int | Tamaño de la partición en bytes |
| `part_name` | char[16] | Nombre de la partición |
| `part_correlative` | int | Correlativo de la partición (-1 inicialmente, incrementa al montar) |
| `part_id` | char[4] | Identificador único asignado al montar |

### Extended Boot Record (EBR)
El EBR describe particiones lógicas dentro de una partición extendida, organizadas en una lista ligada.

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `part_mount` | char | Estado de montaje de la partición |
| `part_fit` | char | Tipo de ajuste (B, F, W) |
| `part_start` | int | Byte inicial de la partición |
| `part_s` | int | Tamaño en bytes |
| `part_next` | int | Posición en bytes del siguiente EBR (-1 si no existe) |
| `part_name` | char[16] | Nombre de la partición |

### Estructuras del Sistema de Archivos EXT2
El sistema de archivos EXT2 se organiza de la siguiente manera:

| Superbloque | Bitmap de Inodos | Bitmap de Bloques | Inodos | Bloques |
|-------------|------------------|------------------|--------|---------|

#### Superbloque
El superbloque almacena metadatos del sistema de archivos.

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `s_filesystem_type` | int | Identificador del sistema de archivos |
| `s_inodes_count` | int | Número total de inodos |
| `s_blocks_count` | int | Número total de bloques |
| `s_free_blocks_count` | int | Número de bloques libres |
| `s_free_inodes_count` | int | Número de inodos libres |
| `s_mtime` | time | Fecha del último montaje |
| `s_umtime` | time | Fecha del último desmontaje |
| `s_mnt_count` | int | Contador de montajes |
| `s_magic` | int | Número mágico (0xEF53) |
| `s_inode_s` | int | Tamaño de cada inodo |
| `s_block_s` | int | Tamaño de cada bloque |
| `s_first_ino` | int | Dirección del primer inodo libre |
| `s_first_blo` | int | Dirección del primer bloque libre |
| `s_bm_inode_start` | int | Inicio del bitmap de inodos |
| `s_bm_block_start` | int | Inicio del bitmap de bloques |
| `s_inode_start` | int | Inicio de la tabla de inodos |
| `s_block_start` | int | Inicio de la tabla de bloques |

#### Inodo
Los inodos contienen metadatos de archivos y carpetas.

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `i_uid` | int | Identificador del propietario (UID) |
| `i_gid` | int | Identificador del grupo (GID) |
| `i_s` | int | Tamaño del archivo en bytes |
| `i_atime` | time | Fecha del último acceso |
| `i_ctime` | time | Fecha de creación |
| `i_mtime` | time | Fecha de la última modificación |
| `i_block` | int[15] | Punteros a bloques (12 directos, 1 indirecto simple, 1 doble indirecto, 1 triple indirecto) |
| `i_type` | char | Tipo (1: Archivo, 0: Carpeta) |
| `i_perm` | char[3] | Permisos UGO en formato octal |

#### Bloques
Los bloques son unidades de 64 bytes para almacenar datos, con tres tipos:

- **Bloque de Carpeta**:
  | Campo | Tipo | Descripción |
  |-------|------|-------------|
  | `b_content` | content[4] | Arreglo de entradas de carpeta |
  - Subestructura `b_content`:
    | Campo | Tipo | Descripción |
    |-------|------|-------------|
    | `b_name` | char[12] | Nombre del archivo o carpeta |
    | `b_inodo` | int | Puntero al inodo correspondiente |

- **Bloque de Archivo**:
  | Campo | Tipo | Descripción |
  |-------|------|-------------|
  | `b_content` | char[64] | Contenido del archivo |

- **Bloque de Punteros**:
  | Campo | Tipo | Descripción |
  |-------|------|-------------|
  | `b_pointers` | int[16] | Arreglo de punteros a otros bloques |

#### Bitmap
- **Bitmap de Inodos**: Registra el estado de los inodos (0: libre, 1: ocupado).
- **Bitmap de Bloques**: Registra el estado de los bloques (0: libre, 1: ocupado).

### Estructuras del Sistema de Archivos EXT3
El sistema EXT3 extiende EXT2 al incorporar un journal para el registro de operaciones.

| Superbloque | Journaling | Bitmap de Inodos | Bitmap de Bloques | Inodos | Bloques |
|-------------|------------|------------------|------------------|--------|---------|

#### Journal
El journal registra todas las operaciones realizadas en el sistema de archivos.

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `count` | int | Contador de entradas en el journal |
| `content` | information | Detalles de la operación |

- Subestructura `information`:
  | Campo | Tipo | Descripción |
  |-------|------|-------------|
  | `i_operation` | char[10] | Tipo de operación realizada |
  | `i_path` | char[32] | Ruta asociada a la operación |
  | `i_content` | char[64] | Contenido del archivo (si aplica) |
  | `i_date` | float | Marca de tiempo de la operación |

#### Cálculo de Inodos y Bloques
Para ambos sistemas (EXT2 y EXT3):
- El número de bloques es igual a 3 veces el número de inodos.
- Fórmula: `tamaño_particion = sizeof(superbloque) + n * sizeof(journaling) + n + 3 * n + n * sizeof(inodos) + 3 * n * sizeof(bloque)`
- `numero_estructuras = floor(n)`
- En EXT3, el journal tiene un tamaño fijo de 50 entradas.

## Comandos Implementados
El sistema implementa un conjunto de comandos para la gestión de discos, particiones, archivos, usuarios, recuperación y generación de reportes. A continuación, se detalla cada comando, incluyendo sus parámetros, funcionalidad y ejemplos.

### Comandos de Gestión de Discos
#### MKDISK
Crea un archivo `.mia` que simula un disco, inicializado con ceros binarios.

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-size` | Requerido | Tamaño del disco (valor mayor a 0) |
| `-fit` | Opcional | Tipo de ajuste (BF: Best Fit, FF: First Fit, WF: Worst Fit; por defecto: FF) |
| `-unit` | Opcional | Unidad de tamaño (K: Kilobytes, M: Megabytes; por defecto: M) |
| `-path` | Requerido | Ruta del archivo (crea directorios superiores si no existen) |

**Ejemplo**:
```
mkdisk -size=3000 -unit=K -path=/home/usuario/Disco1.mia
```
Crea un disco de 3000 KB en la ruta `/home/usuario/Disco1.mia`.

#### RMDISK
Elimina un archivo `.mia`.

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-path` | Requerido | Ruta del archivo del disco |

**Ejemplo**:
```
rmdisk -path=/home/discos/Disco4.mia
```
Elimina el archivo `Disco4.mia`.

#### FDISK
Administra particiones (creación, eliminación y redimensionamiento).

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-size` | Requerido (creación) | Tamaño de la partición (valor mayor a 0) |
| `-unit` | Opcional | Unidad de tamaño (B: Bytes, K: Kilobytes, M: Megabytes; por defecto: K) |
| `-path` | Requerido | Ruta del archivo del disco |
| `-type` | Opcional | Tipo de partición (P: Primaria, E: Extendida, L: Lógica; por defecto: P) |
| `-fit` | Opcional | Tipo de ajuste (BF, FF, WF; por defecto: WF) |
| `-name` | Requerido | Nombre de la partición (único por disco) |
| `-delete` | Opcional | Modo de eliminación (Fast: marca como vacío, Full: llena con ceros) |
| `-add` | Opcional | Añade o elimina espacio (valor positivo o negativo) |

**Ejemplo**:
```
fdisk -size=300 -path=/home/Disco1.mia -name=Particion1
```
Crea una partición primaria de 300 KB llamada `Particion1`.

```
fdisk -delete=fast -name=Particion1 -path=/home/Disco1.mia
```
Elimina `Particion1` en modo rápido.

#### MOUNT
Monta una partición, asignándole un identificador único basado en el carné del estudiante.

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-path` | Requerido | Ruta del archivo del disco |
| `-name` | Requerido | Nombre de la partición |

**Ejemplo**:
```
mount -path=/home/Disco2.mia -name=Part2
```
Monta la partición `Part2` con el identificador `341A` (suponiendo carné `202401234`).

#### MOUNTED
Lista todas las particiones montadas con sus respectivos identificadores.

**Ejemplo**:
```
mounted
```
Salida: `341A, 342A, 341B, 341C`

#### UNMOUNT
Desmonta una partición, restableciendo su correlativo a 0.

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-id` | Requerido | Identificador de la partición |

**Ejemplo**:
```
unmount -id=341A
```
Desmonta la partición con identificador `341A`.

### Comandos de Gestión del Sistema de Archivos
#### MKFS
Formatea una partición en formato EXT2 o EXT3, creando las estructuras necesarias y el archivo `users.txt`.

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-id` | Requerido | Identificador de la partición |
| `-type` | Opcional | Tipo de formateo (Full; por defecto: Full) |
| `-fs` | Opcional | Sistema de archivos (2fs: EXT2, 3fs: EXT3; por defecto: 2fs) |

**Ejemplo**:
```
mkfs -id=061A -fs=3fs
```
Formatea la partición con identificador `061A` en formato EXT3.

#### CAT
Muestra el contenido de archivos si el usuario tiene permisos de lectura.

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-fileN` | Requerido | Rutas de los archivos (soporta múltiples archivos) |

**Ejemplo**:
```
cat -file1=/home/usuario/docs/a.txt
```
Muestra el contenido del archivo `a.txt`.

#### MKFILE
Crea un archivo con contenido o tamaño especificado.

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-path` | Requerido | Ruta del archivo |
| `-r` | Opcional | Crea directorios superiores si no existen |
| `-size` | Opcional | Tamaño del archivo en bytes (rellena con dígitos 0-9) |
| `-cont` | Opcional | Ruta al archivo con el contenido |

**Ejemplo**:
```
mkfile -size=15 -path=/home/usuario/docs/a.txt -r
```
Crea el archivo `a.txt` con 15 bytes de contenido numérico.

#### MKDIR
Crea un directorio.

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-path` | Requerido | Ruta del directorio |
| `-p` | Opcional | Crea directorios superiores si no existen |

**Ejemplo**:
```
mkdir -p -path=/home/usuario/docs/usac
```
Crea el directorio `usac`, incluyendo los directorios superiores.

#### REMOVE
Elimina un archivo o directorio si el usuario tiene permisos de escritura.

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-path` | Requerido | Ruta del archivo o directorio |

**Ejemplo**:
```
remove -path=/home/usuario/docs/a.txt
```
Elimina el archivo `a.txt`.

#### EDIT
Modifica el contenido de un archivo si el usuario tiene permisos de lectura y escritura.

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-path` | Requerido | Ruta del archivo |
| `-contenido` | Requerido | Ruta al archivo con el nuevo contenido |

**Ejemplo**:
```
edit -path=/home/usuario/docs/a.txt -contenido=/root/usuario/files/a.txt
```
Actualiza el contenido de `a.txt` con el nuevo contenido.

#### RENAME
Renombra un archivo o directorio si el usuario tiene permisos de escritura.

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-path` | Requerido | Ruta actual del archivo o directorio |
| `-name` | Requerido | Nuevo nombre |

**Ejemplo**:
```
rename -path=/home/usuario/docs/a.txt -name=b1.txt
```
Renombra `a.txt` a `b1.txt`.

#### COPY
Copia un archivo o directorio a un destino si el usuario tiene permisos de lectura.

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-path` | Requerido | Ruta del archivo o directorio de origen |
| `-destino` | Requerido | Ruta del destino |

**Ejemplo**:
```
copy -path=/home/usuario/documentos -destino=/home/imagenes
```
Copia el directorio `documentos` a la ruta `imagenes`.

#### MOVE
Mueve un archivo o directorio a un destino si el usuario tiene permisos de escritura.

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-path` | Requerido | Ruta del archivo o directorio de origen |
| `-destino` | Requerido | Ruta del destino |

**Ejemplo**:
```
move -path=/home/usuario/documentos -destino=/home/imagenes
```
Mueve el directorio `documentos` a la ruta `imagenes`.

#### FIND
Busca archivos o directorios por nombre, soportando comodines (`?` para un carácter, `*` para múltiples caracteres).

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-path` | Requerido | Ruta del directorio inicial |
| `-name` | Requerido | Patrón de búsqueda |

**Ejemplo**:
```
find -path=/ -name=?.*
```
Lista archivos o directorios con nombres de un carácter.

#### CHOWN
Cambia el propietario de un archivo o directorio (solo permitido para el usuario `root` o el propietario actual).

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-path` | Requerido | Ruta del archivo o directorio |
| `-r` | Opcional | Aplica el cambio recursivamente |
| `-usuario` | Requerido | Nombre del nuevo propietario |

**Ejemplo**:
```
chown -path=/home -r -usuario=usuario2
```
Cambia el propietario de `/home` recursivamente a `usuario2`.

#### CHMOD
Modifica los permisos de un archivo o directorio (solo permitido para el usuario `root`).

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-path` | Requerido | Ruta del archivo o directorio |
| `-ugo` | Requerido | Permisos UGO en formato octal (por ejemplo, 764) |
| `-r` | Opcional | Aplica el cambio recursivamente |

**Ejemplo**:
```
chmod -path=/home -ugo=777
```
Establece los permisos de `/home` a 777.

### Comandos de Gestión de Usuarios y Grupos
#### LOGIN
Inicia sesión de un usuario en una partición.

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-user` | Requerido | Nombre de usuario |
| `-pass` | Requerido | Contraseña |
| `-id` | Requerido | Identificador de la partición |

**Ejemplo**:
```
login -user=root -pass=123 -id=062A
```
Inicia sesión como usuario `root` en la partición `062A`.

#### LOGOUT
Cierra la sesión del usuario activo.

**Ejemplo**:
```
logout
```
Finaliza la sesión actual.

#### MKGRP
Crea un grupo (solo permitido para el usuario `root`).

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-name` | Requerido | Nombre del grupo |

**Ejemplo**:
```
mkgrp -name=usuarios
```
Crea el grupo `usuarios`.

#### RMGRP
Elimina un grupo (solo permitido para el usuario `root`).

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-name` | Requerido | Nombre del grupo |

**Ejemplo**:
```
rmgrp -name=usuarios
```
Elimina el grupo `usuarios`.

#### MKUSR
Crea un usuario (solo permitido para el usuario `root`).

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-user` | Requerido | Nombre de usuario |
| `-pass` | Requerido | Contraseña |
| `-grp` | Requerido | Nombre del grupo |

**Ejemplo**:
```
mkusr -user=usuario1 -pass=usuario -grp=usuarios
```
Crea el usuario `usuario1` en el grupo `usuarios`.

#### RMUSR
Elimina un usuario (solo permitido para el usuario `root`).

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-user` | Requerido | Nombre de usuario |

**Ejemplo**:
```
rmusr -user=usuario1
```
Elimina el usuario `usuario1`.

#### CHGRP
Cambia el grupo al que pertenece un usuario (solo permitido para el usuario `root`).

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-user` | Requerido | Nombre de usuario |
| `-grp` | Requerido | Nombre del nuevo grupo |

**Ejemplo**:
```
chgrp -user=usuario2 -grp=grupo1
```
Cambia el grupo de `usuario2` a `grupo1`.

### Comandos de Recuperación y Journaling
#### RECOVERY
Recupera un sistema de archivos EXT3 utilizando el journal y el superbloque.

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-id` | Requerido | Identificador de la partición |

**Ejemplo**:
```
recovery -id=061Disco1
```
Recupera el sistema de archivos de la partición `061Disco1`.

#### LOSS
Simula la pérdida de datos al llenar con ceros las áreas de inodos y bloques.

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-id` | Requerido | Identificador de la partición |

**Ejemplo**:
```
loss -id=061Disco1
```
Simula la pérdida de datos en la partición `061Disco1`.

#### JOURNALING
Muestra las entradas del journal de una partición EXT3.

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-id` | Requerido | Identificador de la partición |

**Ejemplo**:
```
journaling -id=061Disco1
```
Salida:
| Operación | Ruta | Contenido | Fecha |
|-----------|------|-----------|-------|
| mkdir | / | - | 07/04/2025 19:07 |
| mkfile | /ejemplo.txt | 12345678901234567890 | 07/04/2025 19:07 |

### Comandos de Generación de Reportes
#### REP
Genera reportes utilizando Graphviz.

| Parámetro | Categoría | Descripción |
|-----------|-----------|-------------|
| `-name` | Requerido | Tipo de reporte (mbr, disk, inode, block, bm_inode, bm_block, tree, sb, file, ls) |
| `-path` | Requerido | Ruta del archivo de salida |
| `-id` | Requerido | Identificador de la partición |
| `-path_file_ls` | Opcional | Ruta del archivo o carpeta para reportes `file` o `ls` |

**Ejemplo**:
```
rep -id=A118 -path=/home/usuario/reportes/reporte1.jpg -name=mbr
```
Genera un reporte del MBR.

## Notas de Implementación
- **Permisos**: El sistema implementa permisos UGO, donde el usuario `root` tiene acceso completo (777). Los permisos se verifican para todas las operaciones, salvo aquellas ejecutadas por `root`.
- **Archivo `users.txt`**: Se almacena en la raíz de cada partición y contiene registros de grupos y usuarios en el siguiente formato:
  ```
  GID, Tipo, Grupo \n
  UID, Tipo, Grupo, Usuario, Contraseña \n
  ```
  Contenido inicial: `1, G, root \n 1, U, root, root, 123 \n`
- **Journaling**: El journal de EXT3 registra todas las operaciones, permitiendo la recuperación del sistema a un estado consistente.
- **Gestión de Errores**: Los comandos validan los parámetros y permisos, proporcionando mensajes de error detallados para entradas inválidas o acciones no autorizadas.
