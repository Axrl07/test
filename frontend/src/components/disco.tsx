import diskIcon from "../assets/disk-icon.png";

interface DiskProps {
  disk: {
    ruta: string;
    nombre: string;
    size: string;
    fechaCreacion: string;
    id: number;
    ajuste: string;
    letra: string;
  };
  onClick: () => void;
}

const Disco: React.FC<DiskProps> = ({ disk, onClick }) => {
  return (
    <div
      className="card text-center p-3 border-0 shadow-sm"
      style={{ width: "140px", cursor: "pointer" }}
      onClick={onClick}
    >
      <img src={diskIcon} alt="Disk Icon" className="card-img-top mx-auto" style={{ width: "80px" }} />
      <div className="card-body p-1">
        <p className="card-text small fw-semibold">{disk.nombre}</p>
      </div>
    </div>
  );
};

export default Disco;