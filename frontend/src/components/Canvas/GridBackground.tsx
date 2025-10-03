interface GridBackgroundProps {
  gridSize?: number;
  gridColor?: string;
  canvasWidth?: number;
  canvasHeight?: number;
}

const GridBackground: React.FC<GridBackgroundProps> = ({
  gridSize = 50,
  gridColor = "#e5e5e5",
  canvasWidth = 5000,
  canvasHeight = 5000,
}) => {
  return (
    <>
      {/* Grid pattern */}
      <div
        className="absolute inset-0 pointer-events-none"
        style={{
          backgroundImage: `
            linear-gradient(${gridColor} 1px, transparent 1px),
            linear-gradient(90deg, ${gridColor} 1px, transparent 1px)
          `,
          backgroundSize: `${gridSize}px ${gridSize}px`,
        }}
      />

      {/* Center watermark */}
      <div
        className="absolute pointer-events-none font-custom"
        style={{
          left: `${canvasWidth / 2}px`,
          top: `${canvasHeight / 2}px`,
          transform: "translate(-50%, -50%)",
          fontSize: "100px",
          fontWeight: 300,
          color: "#000000",
          opacity: 0.03,
          userSelect: "none",
          whiteSpace: "nowrap",
        }}
      >
        asocial
      </div>
    </>
  );
};

export default GridBackground;
