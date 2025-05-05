import { useEffect, useState } from "react";
import Archivo from "../components/archivo";
import Carpeta from "../components/carpeta";
import { CarpetaInterface, ArchivoInterface, ObtenerSistema } from "../services/api";
import { useParams } from "react-router-dom";
import InformacionArchivo from "../components/archivoInfo";
import InformacionCarpeta from "../components/carpetaInfo";

function ExploradorArchivos() {
    const [historial, setHistorial] = useState<CarpetaInterface[]>([]);
    const [archivosYCarpetas, setArchivosYCarpetas] = useState<(CarpetaInterface | ArchivoInterface)[]>([]);
    const [selectedItem, setSelectedItem] = useState<ArchivoInterface | CarpetaInterface | null>(null);
    const { idPart } = useParams<{ idPart: string }>();
    const [loading, setLoading] = useState(true);

    const abrirCarpeta = (carpeta: CarpetaInterface) => {
        setHistorial((prev) => [...prev, carpeta]); // Guardamos la carpeta actual en el historial
        setArchivosYCarpetas(carpeta.hijos || []);
        setSelectedItem(null); // Limpia selección al cambiar de carpeta
    };

    const regresar = () => {
        if (historial.length > 1) {
            const nuevoHistorial = [...historial];
            nuevoHistorial.pop(); // quitamos la actual
            const carpetaAnterior = nuevoHistorial[nuevoHistorial.length - 1];
            setHistorial(nuevoHistorial);
            setArchivosYCarpetas(carpetaAnterior.hijos || []);
            setSelectedItem(null);
        } else {
            // Si solo hay una carpeta, regresa a la raíz
            handlerArchivosYCarpetas(idPart!);
            setHistorial([]);
        }
    };

    const handlerArchivosYCarpetas = async (id: string) => {
        setLoading(true);
        try {
            const respuesta = await ObtenerSistema(id);
            const sistema: CarpetaInterface = respuesta.sistema; // ya es un objeto, no un string
            setArchivosYCarpetas(sistema.hijos || []);
        } catch (error) {
            console.error("Error al obtener archivos y carpetas:", error);
            setArchivosYCarpetas([]);
        } finally {
            setLoading(false);
        }
    };

    function esCarpeta(item: CarpetaInterface | ArchivoInterface): item is CarpetaInterface {
        return item.tipo === "carpeta";
    }

    useEffect(() => {
        if (idPart) {
            (async () => {
                setLoading(true);
                try {
                    const respuesta = await ObtenerSistema(idPart);
                    const sistema: CarpetaInterface = respuesta.sistema;
                    setHistorial([sistema]);
                    setArchivosYCarpetas(sistema.hijos || []);
                } catch (error) {
                    console.error("Error al obtener sistema:", error);
                    setHistorial([]);
                    setArchivosYCarpetas([]);
                } finally {
                    setLoading(false);
                }
            })();
        }
    }, [idPart]);

    return (
        <div className="container py-5">
            {/* Barra de búsqueda / breadcrumb */}
            <div className="d-flex align-items-center mb-3">
                <div
                    className="breadcrumb w-100 p-3 rounded-3 shadow-sm bg-light"
                    style={{
                        fontSize: "1.1rem",
                        display: "flex",
                        flexWrap: "wrap",
                        alignItems: "center",
                        justifyContent: "flex-start",
                    }}
                >
                    {historial.map((carpeta, index) => (
                        <span key={index} className="d-flex align-items-center">
                            {/* Mostrar solo si el nombre de la carpeta no es '/' */}
                            {<span>{carpeta.nombre}</span>}
                            {index < historial.length - 1 && carpeta.nombre !== "/" && " / "}
                        </span>
                    ))}
                </div>
            </div>

            <div className="d-flex justify-content-end gap-2 mb-3">
                <button className="btn btn-warning" onClick={regresar} disabled={historial.length <= 1}>
                    Regresar
                </button>
            </div>

            <div className="row g-4">
                {/* Panel izquierdo: 60% */}
                <div className="col-md-7">
                    <div className="card shadow-sm">
                        <div className="card-body">
                            <h2 className="card-title text-center mb-4">Explorador de Archivos</h2>
                            <div className="d-flex flex-wrap gap-3 justify-content-center">
                                {loading ? (
                                    <p className="text-muted">Cargando archivos y carpetas...</p>
                                ) : archivosYCarpetas.length === 0 ? (
                                    <p className="text-muted">No hay archivos o carpetas disponibles.</p>
                                ) : (
                                    archivosYCarpetas.map((item, index) => (
                                        esCarpeta(item) ? (
                                            <Carpeta
                                                key={index}
                                                folder={item}
                                                onClick={() => setSelectedItem(item)}
                                            />
                                        ) : (
                                            <Archivo
                                                key={index}
                                                file={item}
                                                onClick={() => setSelectedItem(item)}
                                            />
                                        )
                                    ))
                                )}
                            </div>
                        </div>
                    </div>
                </div>

                {/* Panel derecho: 40% */}
                <div className="col-md-5">
                    <div className="card shadow-sm h-100">
                        <div className="card-body">
                            {selectedItem ? (
                                <>
                                    {'contenido' in selectedItem ? (
                                        <InformacionArchivo archivo={selectedItem as ArchivoInterface} />
                                    ) : (
                                        <InformacionCarpeta
                                            carpeta={selectedItem as CarpetaInterface}
                                            onAbrir={abrirCarpeta}
                                        />
                                    )}
                                </>
                            ) : (
                                <p className="text-muted text-center mt-4">
                                    Selecciona un archivo o carpeta para ver más información.
                                </p>
                            )}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default ExploradorArchivos;
