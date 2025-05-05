import { /*ArchivoInterface ,*/ CarpetaInterface } from "../services/api";
// import Carpeta from "./carpeta";
// import Archivo from "./archivo";

interface CarpetaProps {
  carpeta: CarpetaInterface;
  onAbrir: (carpeta: CarpetaInterface) => void; // nueva prop
}

const InformacionCarpeta: React.FC<CarpetaProps> = ({ carpeta, onAbrir }) => {
  // function esCarpeta(item: CarpetaInterface | ArchivoInterface): item is CarpetaInterface {
  //         return item.tipo === "carpeta";
  //     }

  return (
    <div>
      <h3 className="fw-bold mb-3 text-center">Información de la Carpeta</h3>
      <ul className="list-group list-group-flush">
        <li className="list-group-item"><strong>Nombre:</strong> {carpeta.nombre}</li>
        <li className="list-group-item"><strong>Permisos:</strong> {carpeta.permisos}</li>
        <li className="list-group-item"><strong>Propietario:</strong> {carpeta.propietario}</li>
        <li className="list-group-item"><strong>Grupo:</strong> {carpeta.grupo}</li>
        <li className="list-group-item"><strong>Fecha de Creación:</strong> {carpeta.fechaCreacion}</li>
        <li className="list-group-item"><strong>Fecha de Modificación:</strong> {carpeta.fechaModificacion}</li>
        <li className="list-group-item"><strong>Fecha de Acceso:</strong> {carpeta.fechaAcceso}</li>
      </ul>

      <div className="d-grid gap-2 mt-3">
        <button className="btn btn-primary" onClick={() => onAbrir(carpeta)}>
          Abrir
        </button>
      </div>

      {/* <h4 className="mt-4">Contenido de la Carpeta:</h4>
      <div className="d-flex flex-wrap gap-3">
        {carpeta.hijos.map((item, index) =>
          esCarpeta(item) ? (
            <Carpeta key={index} folder={item} onClick={() => {}} />
          ) : (
            <Archivo key={index} file={item} onClick={() => {}} />
          )
        )}
      </div> */}
    </div>
  );
};

export default InformacionCarpeta;
