import partitionIcon from "../assets/partition-icon.png";

interface PartProps {
  part: {
    estado: string,
    tipo: string,
    ajuste: string,
    inicio: string,
    size: string,
    nombre: string,
    correlativo: string,
    id: string,
  };
  onClick: () => void;
}

const Particion: React.FC<PartProps> = ({ part, onClick }) => {
  return (
    <div
      className="card text-center p-3 border-0 shadow-sm"
      style={{ width: "140px", cursor: "pointer" }}
      onClick={onClick}
    >
      <img src={partitionIcon} alt="Part Icon" className="card-img-top mx-auto" style={{ width: "80px" }} />
      <div className="card-body p-1">
        <p className="card-text small fw-semibold">{part.nombre}</p>
      </div>
    </div>
  );
};

export default Particion;