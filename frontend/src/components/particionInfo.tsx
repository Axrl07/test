
import { useNavigate } from "react-router-dom";
import { PartInterface } from "../services/api";

interface PartProps {
    part: PartInterface;
}

const InformacionParticion: React.FC<PartProps> = ({ part }) => {
    const navigate = useNavigate();
    const handleParticion = () => {
        navigate(`/${part.id}/sistema`)
    };
    return (
        <div>
            <h3 className="fw-bold mb-3 text-center">Información de la Partición</h3>
            <ul className="list-group list-group-flush">
                <li className="list-group-item"><strong>Nombre:</strong> {part.nombre}</li>
                <li className="list-group-item"><strong>Estado:</strong> {part.estado}</li>
                <li className="list-group-item"><strong>Tipo:</strong> {part.tipo}</li>
                <li className="list-group-item"><strong>Ajuste:</strong> {part.ajuste}</li>
                <li className="list-group-item"><strong>Inicio:</strong> {part.inicio}</li>
                <li className="list-group-item"><strong>Tamaño:</strong> {part.size}</li>
                {
                    part.id != "0" ? (
                        <li className="list-group-item"><strong>ID:</strong> {part.id}</li>
                    ) : (
                        <></>
                    )
                }
                {
                    part.correlativo != "-1" ? (
                        <li className="list-group-item"><strong>Correlativo:</strong> {part.correlativo}</li>
                    ) : (
                        <li className="list-group-item"></li>
                    )
                }
            </ul>

            <div className="d-grid gap-2">
                {
                    part.correlativo != "-1" ? (
                        <button className="btn btn-primary mt-3" onClick={handleParticion}>
                            Ingresar
                        </button>
                    ) : (
                        <div className="alert alert-primary d-flex align-items-center" role="alert">
                            <i className="bi bi-exclamation-triangle-fill flex-shrink-0 me-2" role="img" aria-label="Warning:"></i>
                            <div>
                                No se puede ingresar a particiones que no están montadas o no son primarias.
                            </div>
                        </div>
                    )
                }
            </div>
        </div>
    );
};

export default InformacionParticion;
