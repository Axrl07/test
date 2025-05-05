
import { useNavigate } from "react-router-dom";
import { diskInterface } from "../services/api";

interface DiskProps {
  disk: diskInterface;
}

const InformacionDisco: React.FC<DiskProps> = ({ disk }) => {
  const navigate = useNavigate();
  
  const handleParticion =  () => {
    navigate(`/${disk.nombre}/particiones`);
  };

  return (
    <div>
      <h3 className="fw-bold mb-3 text-center">Información del Disco</h3>
      <ul className="list-group list-group-flush">
        <li className="list-group-item"><strong>Nombre:</strong> {disk.nombre}</li>
        <li className="list-group-item"><strong>Ruta:</strong> {disk.ruta}</li>
        <li className="list-group-item"><strong>Tamaño:</strong> {disk.size}</li>
        <li className="list-group-item"><strong>Fecha de creación:</strong> {disk.fechaCreacion}</li>
        <li className="list-group-item"><strong>ID:</strong> {disk.id}</li>
        <li className="list-group-item"><strong>Ajuste:</strong> {disk.ajuste}</li>
        <li className="list-group-item"><strong>Letra:</strong> {disk.letra}</li>
      </ul>
      <div className="d-grid gap-2">
        <button className="btn btn-primary mt-3" onClick={handleParticion}>
          Ingresar
        </button>
      </div>
    </div>
  );
};

export default InformacionDisco;

