interface GridBackgroundProps {
  gridSize?: number;
  gridColor?: string;
}

const GridBackground: React.FC<GridBackgroundProps> = ({
  gridSize = 50,
  gridColor = "#e5e5e5",
}) => {
  return (
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
  );
};

export default GridBackground;
