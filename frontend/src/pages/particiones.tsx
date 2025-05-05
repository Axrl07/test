import { useEffect, useState } from "react";
import Particion from "../components/particion";
import InformacionParticion from "../components/particionInfo";
import { PartInterface, ObtenerParticiones } from "../services/api";
import { useNavigate, useParams } from "react-router";

function Particiones() {
    const [particiones, setParticiones] = useState<PartInterface[]>([]);
    const [selectedPart, setSelectedPart] = useState<PartInterface | null>(null);
    const [loading, setLoading] = useState(true);
    const navigate = useNavigate();
    // obtenemos el nombre del disco
    const { nombreDisco } = useParams<{ nombreDisco: string }>();

    const handleParticiones = async () => {
        setLoading(true);
        try {
            if (nombreDisco) {
                console.log(nombreDisco)
                const respuesta = await ObtenerParticiones(nombreDisco);
                setParticiones(respuesta || []);
            } else {
                setParticiones([]);
            }
        } catch (error) {
            console.error("Error al obtener las particiones:", error);
            setParticiones([]);
        } finally {
            setLoading(false);
        }
    };

    const handleRegresar = () => {
        navigate("/discos");
    };

    useEffect(() => {
        handleParticiones();
    }, []);

    return (
        <div className="container py-5">
            <div className="d-flex justify-content-end gap-2 mb-3">
                <button className="btn btn-warning" onClick={handleRegresar}>Regresar</button>
            </div>

            <div className="row g-4">
                {/* Panel izquierdo */}
                <div className="col-md-7">
                    <div className="card shadow-sm">
                        <div className="card-body">
                            <h2 className="card-title text-center mb-4">Particiones</h2>
                            <div className="d-flex flex-wrap gap-3 justify-content-center">
                                {loading ? (
                                    <p className="text-muted">Cargando particiones...</p>
                                ) : particiones.length === 0 ? (
                                    <p className="text-muted">No hay particiones primarias creadas o montadas actualmente.</p>
                                ) : (
                                    particiones.map((part) => (
                                        <Particion
                                            key={part.nombre}
                                            part={part}
                                            onClick={() => setSelectedPart(part)}
                                        />
                                    ))
                                )}
                            </div>
                        </div>
                    </div>
                </div>

                {/* Panel derecho */}
                <div className="col-md-5">
                    <div className="card shadow-sm h-100">
                        <div className="card-body">
                            {selectedPart ? (
                                <InformacionParticion part={selectedPart} />
                            ) : (

                                <div>
                                <h3 className="fw-bold mb-3 text-center">Información de la Partición</h3>
                                    <p className="text-muted text-center mt-4">
                                        Selecciona una partición para ver la información.
                                    </p>
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default Particiones;
