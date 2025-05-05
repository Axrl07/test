
const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8000";

// interpretación de comandos
export const interpretarComandos = async (comandos: string): Promise<string> => {
  try {
    // hacemos la solicitud al backend oara analizar el comando
    const response = await fetch(`${API_URL}/analizar`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ entrada: comandos }),
    });
    // obtenemos la salida del backend
    const data = await response.json();
    if (!response.ok) {
      throw new Error(`ERROR: respuesta del servidor: ${data.salida}`);
    }
    // retornamos la consola
    return data.salida;
  } catch (error) {
    // lanzamos la excepción
    throw new Error(`ERROR: no ha sido posible ejecutar los comandos por: ${error}`);
  }
};

// para la verificación de usuarios
export interface usuario {
  nombre: string;
  clave: string;
  idParticion: string;
}
export const VerificandoSesionActual = async (): Promise<usuario> => {
  try {
    const loginResponse = await fetch(`${API_URL}/verificarLogin`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    });

    const data = await loginResponse.json();

    if (!loginResponse.ok) {
      throw new Error(`ERROR: no hay sesión iniciada`);
    }

    return data;
  } catch (error) {
    console.error("Error en verificandoLogin:", error);
    throw error;
  }
};

// para iniciar sesión
export const IniciarSesion = async (nombre: string, clave: string, idPart: string): Promise<string> => {
  try {
    const loginData = {
      nombre: nombre,
      clave: clave,
      idParticion: idPart
    };

    // Hacemos la solicitud al backend para autenticar al usuario
    const loginResponse = await fetch(`${API_URL}/login`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(loginData),
    });

    const loginDataResponse = await loginResponse.json();

    if (!loginResponse.ok) {
      throw loginDataResponse.salida
    }
    return loginDataResponse.salida
  } catch (error) {
    throw new Error(`${error}`);
  }
};

// para cerrar sesión
export const CerrarSesion = async (): Promise<string> => {
  try {
    // Hacemos la solicitud al backend para autenticar al usuario
    const loginResponse = await fetch(`${API_URL}/logout`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    });

    const loginDataResponse = await loginResponse.json();

    if (!loginResponse.ok) {
      throw loginDataResponse.salida
    }
    return loginDataResponse.salida
  } catch (error) {
    throw new Error(`${error}`);
  }
};

// para obtener la tabla journaling
export const ObtenerTablaJournaling = async (): Promise<string> => {
  try {
    // Hacemos la solicitud al backend para autenticar al usuario
    const tablaResponse = await fetch(`${API_URL}/tablaJournaling`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    });

    const loginDataResponse = await tablaResponse.json();

    if (!tablaResponse.ok) {
      throw loginDataResponse.salida
    }
    return loginDataResponse.salida
  } catch (error) {
    throw new Error(`${error}`);
  }
}

// para obtener los discos
export interface diskInterface {
  ruta: string;
  nombre: string;
  size: string;
  fechaCreacion: string;
  id: number;
  ajuste: string;
  letra: string;
}
export const ObtenerDiscos = async (): Promise<diskInterface[]> => {
  try {
    const discosResponse = await fetch(`${API_URL}/discos`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    });

    const response = await discosResponse.json();
    
    if (!discosResponse.ok) {
      throw response.salida;
    }
    return response; // <- aquí asumimos que esto es diskInterface[]
  } catch (error) {
    throw new Error(`${error}`);
  }
};


// para obtener las particiones
export interface PartInterface {
  estado: string,
  tipo: string,
  ajuste: string,
  inicio: string,
  size: string,
  nombre: string,
  correlativo: string,
  id: string,
}
export const ObtenerParticiones = async (discoNombre: string): Promise<PartInterface[]> => {
  try {
    const discosResponse = await fetch(`${API_URL}/particiones`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({entrada: discoNombre}),
    });

    const response = await discosResponse.json(); 
    if (!discosResponse.ok) {
      throw response.salida;
    }
    return response; // <- aquí asumimos que esto es PartInterface[]
  } catch (error) {
    throw new Error(`${error}`);
  }
}

// para obtener el sistema de archivos
export interface ArchivoInterface {
  nombre: string;
  permisos: string;
  propietario: string;
  grupo: string;
  fechaCreacion: string;
  fechaModificacion: string;
  fechaAcceso: string;
  tipo: string;
  contenido: string;
  size: number;
}
export interface CarpetaInterface {
  nombre: string;
  permisos: string;
  propietario: string;
  grupo: string;
  fechaCreacion: string;
  fechaModificacion: string;
  fechaAcceso: string;
  tipo: string;
  hijos: (ArchivoInterface | CarpetaInterface)[];
}

type SistemaRespuesta = {
  sistema: CarpetaInterface;
  mensaje: string;
};
export const ObtenerSistema = async (idPart: string): Promise<SistemaRespuesta> => {
  try {
    const discosResponse = await fetch(`${API_URL}/sistema`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({entrada: idPart}),
    });

    const response = await discosResponse.json(); 
    if (!discosResponse.ok) {
      throw response.salida;
    }
    return response; // <- aquí asumimos que esto es PartInterface[]
  } catch (error) {
    throw new Error(`${error}`);
  }
}