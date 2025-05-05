# MIA_1S_P2_202209714
Sistema de Archivos Ext2 y Ext3 online

# Instalaciones
Tomar en cuenta que estas operaciones se realizaron en el sistema Linux, por lo que las instrucciones de instalación aseguran su efectividad solamente dentro de este sistema.  

## Graphviz
Para instalar graphviz ejecute los siguientes comando (le solicitará la contraseña ya que requiere permisos para instalarse):  
```bash
sudo apt update
```  
```bash
sudo apt install graphviz
```  
Por último verificamos la instalación con:  
```bash
dot -v
```  

## Go
1. vaya al sitio oficial de go *https://go.dev/dl/* y descargue la versión más reciente, actualmente es **go1.24.2.linux-amd64.tar.gz**, tome en cuenta que si descarga otra versión debe cambiar el nombre del archivo en el paso 2. 

2. Para eliminar cualquier instalación anterior de Go , alojada en `/usr/local/go` (en caso de que exista). Extraiga el archivo que acaba de descargar en `/usr/local`, luego ejecute los siguientes comandos: 
    ```bash
    rm -rf /usr/local/go    // omitase si no existe la carpeta
    tar -C /usr/local -xzf go1.24.2.linux-amd64.de.gz
    ```
    Nota: Recuerde que si la versión que descargó es diferente debe cambiar el nombre **go1.24.2.linux-amd64.de.gz** por el del archivo que descargó. 


3. Agregamos el PATH de Go a las variables de entorno ejecutando el siguiente comando: 
    ```bash
    nano ~/.bashrc
    ``` 
    Una vez adentro del documento vamos a bajar hasta el final e ingresaremos las siguientes lineas (para pegar dentro del documento usamos **Ctrl+Shift+V**): 
    ```bash
    # Go - configuración del PATH
    export PATH=$PATH:/usr/local/go/bin
    export GOPATH=$HOME/go
    export PATH=$PATH:$GOPATH/bin
    ```

4. Finalmente comprobamos la versión de Go que se ha instalado con:
    ```bash
    go version
    ```

## Air  
Air es un framework que se encarga de refrescar el servidor (bajarlo, cargar cambios y volverlo a levantar) cada que se realice un cambio en el backend. Para instalarlo con Go (forma recomendada) necesitamos como mínimo la versión 1.23 de Go instalada.
```bash
go install github.com/air-verse/air@latest
```

## Node.js
Para este proyecto se instaló la versión **v22.14.0(LTS)** usando **nvm** con **npm**  

1. Descargar e instalar **nvm**  
    ```bash
    curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.2/install.sh | bash
    ```  
2. En lugar de reiniciar el shell  
    ```bash
    \. "$HOME/.nvm/nvm.sh"
    ```  

3. Descargando e instalando **Node.js**  
    ```bash
    nvm install 22
    ```  

4. Verificando las versiones de **Node.js**, **nvm** y **npm**  
    ```bash
    node -v
    nvm current
    npm -v
    ```  

## React con Vite
React es el encargado de toda la parte de mi frontend, por lo que para que funcione debemos instalar las librerías necesarias, para ello debemos dirigirnos a la carpeta `/frontend` y ejecutar el comando:  
```bash
npm install
```  

# Agregando tecnologías al proyecto

## Fiber
fiber literalmente es el servidor que se levanta para recibir las peticiones realizadas desde el frontend. Para agregarlo al proyecto debemos dirigirnos a la carpeta `/backend` y ejecutar el comando:  
```bash
go get github.com/gofiber/fiber/v2
```  

# Inicialización

## Backend
Para inicializar el backend debemos dirigirnos a la carpeta `/backend` para ejecutar los comandos.  

primero inicializar Air con:
```bash
air init
```  

Y luego vamos a levantar el servidor con:
```bash
air
```  

## Frontend
Para inicializar el frontend debemos dirigirnos a la carpeta `/frontend` para ejecutar los comandos.  

Ejecutamos React con vite con el comando:
```bash
npm run dev
```  

Luego de ver un mensaje en consola que le muestre el puerto, presionaremos Ctrl + click sobre el enlace con la URL del frontend.