import { useEffect, useState } from "react";
import Disco from "../components/disco";
import InformacionDisco from "../components/discoInfo";
import { diskInterface, ObtenerDiscos } from "../services/api";

function Discos() {
    const [discos, setDiscos] = useState<diskInterface[]>([]);
    const [selectedDisk, setSelectedDisk] = useState<diskInterface | null>(null);
    const [loading, setLoading] = useState(true);

    const handlerDiscos = async () => {
        setLoading(true);
        try {
            const respuesta = await ObtenerDiscos();
            console.log("Discos obtenidos:", respuesta);
            setDiscos(respuesta || []);
        } catch (error) {
            console.error("Error al obtener los discos:", error);
            setDiscos([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        handlerDiscos();
    }, []);

    return (
        <div className="container py-5">
            <div className="row g-4">
                {/* Panel izquierdo: 60% */}
                <div className="col-md-7">
                    <div className="card shadow-sm">
                        <div className="card-body">
                            <h2 className="card-title text-center mb-4">Discos</h2>
                            <div className="d-flex flex-wrap gap-3 justify-content-center">
                                {loading ? (
                                    <p className="text-muted">Cargando discos...</p>
                                ) : discos.length === 0 ? (
                                    <p className="text-muted">No hay discos creados actualmente.</p>
                                ) : (
                                    discos.map((disk) => (
                                        <Disco
                                            key={disk.id}
                                            disk={disk}
                                            onClick={() => setSelectedDisk(disk)}
                                        />
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
                            {selectedDisk ? (
                                <InformacionDisco disk={selectedDisk} />
                            ) : (
                                <p className="text-muted text-center mt-4">
                                    Selecciona un disco para ver la información.
                                </p>
                            )}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default Discos;

