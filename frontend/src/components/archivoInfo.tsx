import { ArchivoInterface } from "../services/api";

interface ArchivoProps {
  archivo: ArchivoInterface;
}

const InformacionArchivo: React.FC<ArchivoProps> = ({ archivo }) => {
  return (
    <div>
      <h3 className="fw-bold mb-3 text-center">Información del Archivo</h3>
      <ul className="list-group list-group-flush">
        <li className="list-group-item"><strong>Nombre:</strong> {archivo.nombre}</li>
        <li className="list-group-item"><strong>Permisos:</strong> {archivo.permisos}</li>
        <li className="list-group-item"><strong>Propietario:</strong> {archivo.propietario}</li>
        <li className="list-group-item"><strong>Grupo:</strong> {archivo.grupo}</li>
        <li className="list-group-item"><strong>Fecha de Creación:</strong> {archivo.fechaCreacion}</li>
        <li className="list-group-item"><strong>Fecha de Modificación:</strong> {archivo.fechaModificacion}</li>
        <li className="list-group-item"><strong>Fecha de Acceso:</strong> {archivo.fechaAcceso}</li>
        <li className="list-group-item"><strong>Tamaño:</strong> {archivo.size} bytes</li>
      </ul>
      <h4 className="mt-3">Contenido del archivo:</h4>
      <pre>{archivo.contenido}</pre>
    </div>
  );
};

export default InformacionArchivo;
