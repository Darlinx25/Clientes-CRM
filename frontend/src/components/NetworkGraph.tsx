import { useRef, useCallback, useMemo, useEffect, useState } from 'react';
import ForceGraph2D from 'react-force-graph-2d';
import { forceX, forceY } from 'd3-force';
import { useTheme, Box, Typography, useMediaQuery } from '@mui/material';
import { useTranslation } from 'react-i18next';
import { GraphData, GraphNode, GraphEdge } from '../types/graph';

interface NetworkGraphProps {
  data: GraphData;
  onNodeClick: (node: GraphNode) => void;
  onActivityClick?: (node: GraphNode) => void;
  showRelationships: boolean;
  showActivities: boolean;
  centeredNodeId?: string;
}

interface ForceGraphData {
  nodes: GraphNode[];
  links: GraphEdge[];
}

const getNodeSize = (type: GraphNode['type']): number => {
  if (type === 'contact') return 12;
  return 6;
};

export default function NetworkGraph({
  data,
  onNodeClick,
  onActivityClick,
  showRelationships,
  showActivities,
  centeredNodeId,
}: NetworkGraphProps) {
  const { t } = useTranslation();
  const theme = useTheme();
  const containerRef = useRef<HTMLDivElement>(null);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const graphRef = useRef<any>(null);
  const [dimensions, setDimensions] = useState({ width: 800, height: 600 });
  const [hoveredEdge, setHoveredEdge] = useState<GraphEdge | null>(null);
  const [hoveredNode, setHoveredNode] = useState<GraphNode | null>(null);
  const [tooltipPos, setTooltipPos] = useState({ x: 0, y: 0 });
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  // Colors from theme
  const relationshipColor = theme.palette.primary.main;
  const activityColor = theme.palette.secondary.main;
  const nodeColor = theme.palette.primary.main;
  const activityNodeColor = theme.palette.secondary.main;
  const textColor = theme.palette.text.primary;
  const bgColor = theme.palette.background.paper;

  // Handle container resize and mouse tracking
  useEffect(() => {
    const updateDimensions = () => {
      if (containerRef.current) {
        const { width, height } = containerRef.current.getBoundingClientRect();
        setDimensions({ width, height });
      }
    };

    const handleMouseMove = (e: MouseEvent) => {
      setTooltipPos({ x: e.clientX, y: e.clientY });
    };

    updateDimensions();
    window.addEventListener('resize', updateDimensions);
    window.addEventListener('mousemove', handleMouseMove);
    return () => {
      window.removeEventListener('resize', updateDimensions);
      window.removeEventListener('mousemove', handleMouseMove);
    };
  }, []);

  // Filter and transform data for the graph
  const graphData: ForceGraphData = useMemo(() => {
    let filteredNodes = data.nodes;

    // Hide activity nodes when the activities toggle is off
    if (!showActivities) {
      filteredNodes = filteredNodes.filter(n => n.type !== 'activity');
    }

    const nodeIds = new Set(filteredNodes.map(n => n.id));

    // Filter edges based on visibility toggles and filtered nodes
    let filteredEdges = data.edges.filter(e => {
      const sourceId = typeof e.source === 'string' ? e.source : e.source.id;
      const targetId = typeof e.target === 'string' ? e.target : e.target.id;

      if (!nodeIds.has(sourceId) || !nodeIds.has(targetId)) return false;
      if (e.type === 'relationship' && !showRelationships) return false;
      return true;
    });

    if (centeredNodeId) {
      const directNeighbors = new Set<string>([centeredNodeId]);

      filteredEdges.forEach(e => {
        const srcId = typeof e.source === 'string' ? e.source : e.source.id;
        const tgtId = typeof e.target === 'string' ? e.target : e.target.id;
        if (srcId === centeredNodeId) directNeighbors.add(tgtId);
        if (tgtId === centeredNodeId) directNeighbors.add(srcId);
      });

      filteredNodes = filteredNodes.filter(n => directNeighbors.has(n.id));
      filteredEdges = filteredEdges.filter(e => {
        const srcId = typeof e.source === 'string' ? e.source : e.source.id;
        const tgtId = typeof e.target === 'string' ? e.target : e.target.id;
        return directNeighbors.has(srcId) && directNeighbors.has(tgtId);
      });
    }

    return {
      nodes: filteredNodes,
      links: filteredEdges,
    };
  }, [data, showRelationships, showActivities, centeredNodeId]);

  // Center and zoom to selected node when centeredNodeId changes
  useEffect(() => {
    if (!centeredNodeId || !graphRef.current) return;

    const node = graphData.nodes.find(n => n.id === centeredNodeId);
    if (!node || node.x == null || node.y == null) return;

    graphRef.current.centerAt(node.x, node.y, 800);
    graphRef.current.zoom(2.5, 800);
  }, [centeredNodeId, graphData.nodes]);

  // Get initials from a name
  const getInitials = (label: string): string => {
    const parts = label.split(' ').filter(Boolean);
    if (parts.length >= 2) {
      return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
    }
    return label.substring(0, 2).toUpperCase();
  };

  // Custom node rendering
  const nodeCanvasObject = useCallback((node: GraphNode, ctx: CanvasRenderingContext2D, globalScale: number) => {
    const isContact = node.type === 'contact';
    const isActivity = node.type === 'activity';
    const size = getNodeSize(node.type);
    const fontSize = Math.max(10 / globalScale, 3);
    const isCentered = node.id === centeredNodeId;

    // Draw highlight ring for centered contact
    if (isCentered) {
      ctx.beginPath();
      ctx.arc(node.x || 0, node.y || 0, size + 5, 0, 2 * Math.PI);
      ctx.strokeStyle = theme.palette.primary.light;
      ctx.lineWidth = 3 / globalScale;
      ctx.stroke();
    }

    // Draw node circle
    ctx.beginPath();
    ctx.arc(node.x || 0, node.y || 0, size, 0, 2 * Math.PI);
    if (isContact) {
      ctx.fillStyle = nodeColor;
    } else {
      ctx.fillStyle = activityNodeColor;
    }
    ctx.fill();

    // Draw border
    ctx.strokeStyle = bgColor;
    ctx.lineWidth = 2 / globalScale;
    ctx.stroke();

    // Draw initials for contacts
    if (isContact && globalScale > 0.5) {
      ctx.font = `bold ${fontSize * 1.2}px Helvetica, Helvetica Neue, Roboto, Arial, sans-serif`;

      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillStyle = '#FFFFFF';
      ctx.fillText(getInitials(node.label), node.x || 0, node.y || 0);
    }

    // Draw label below node when zoomed in enough
    if (globalScale > 0.6 && (isContact || isActivity)) {
      ctx.font = `${fontSize}px Helvetica, Helvetica Neue, Roboto, Arial, sans-serif`;
      ctx.textAlign = 'center';
      ctx.textBaseline = 'top';
      ctx.fillStyle = textColor;
      ctx.fillText(node.label, node.x || 0, (node.y || 0) + size + 4);
    }
  }, [nodeColor, activityNodeColor, bgColor, textColor, centeredNodeId, theme.palette.primary.light]);

  // Custom link rendering
  const linkColor = useCallback((link: GraphEdge) => {
    if (link.type === 'relationship') return relationshipColor;
    return activityColor;
  }, [relationshipColor, activityColor]);

  // Handle node hover
  const handleNodeHover = useCallback((node: GraphNode | null) => {
    setHoveredNode(node);
  }, []);

  // Handle link hover
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handleLinkHover = useCallback((link: any) => {
    setHoveredEdge(link as GraphEdge | null);
  }, []);

  // Handle node click
  const handleNodeClick = useCallback((node: GraphNode) => {
    if (node.type === 'contact') {
      onNodeClick(node);
    } else if (node.type === 'activity') {
      onActivityClick?.(node);
    }
  }, [onNodeClick, onActivityClick]);

  // Configure forces to prevent isolated nodes from drifting too far
  useEffect(() => {
    if (graphRef.current) {
      const fg = graphRef.current;
      fg.d3Force('x', forceX(0).strength(0.05));
      fg.d3Force('y', forceY(0).strength(0.05));
      fg.d3Force('charge')?.strength(-100);
    }
  }, []);

  // Zoom to fit on initial load
  useEffect(() => {
    if (graphRef.current && graphData.nodes.length > 0) {
      setTimeout(() => {
        graphRef.current?.zoomToFit(400, isMobile ? 50 : 80);
      }, 500);
    }
  }, [graphData.nodes.length, isMobile]);

  const getEdgeTypeLabel = (type: string) => {
    if (type === 'relationship') return t('network.legend.relationships');
    return t('network.legend.activities');
  };

  return (
    <Box ref={containerRef} sx={{ width: '100%', height: '100%', position: 'relative' }}>
      <ForceGraph2D
        ref={graphRef}
        width={dimensions.width}
        height={dimensions.height}
        graphData={graphData}
        nodeCanvasObject={nodeCanvasObject}
        nodePointerAreaPaint={(node: GraphNode, color, ctx) => {
          const size = getNodeSize(node.type);
          ctx.beginPath();
          ctx.arc(node.x || 0, node.y || 0, size + 4, 0, 2 * Math.PI);
          ctx.fillStyle = color;
          ctx.fill();
        }}
        linkColor={linkColor}
        linkWidth={2}
        linkDirectionalArrowLength={0}
        onNodeClick={handleNodeClick}
        onNodeHover={handleNodeHover}
        onLinkHover={handleLinkHover}
        cooldownTicks={100}
        enableNodeDrag={true}
        enableZoomInteraction={true}
        enablePanInteraction={true}
        backgroundColor={bgColor}
        nodeId="id"
        linkSource="source"
        linkTarget="target"
      />

      {/* Node / edge tooltip */}
      {(hoveredNode || hoveredEdge) && (
        <Box
          sx={{
            position: 'fixed',
            left: tooltipPos.x + 10,
            top: tooltipPos.y + 10,
            bgcolor: 'background.paper',
            border: 1,
            borderColor: 'divider',
            borderRadius: 1,
            px: 1.5,
            py: 0.75,
            boxShadow: 2,
            pointerEvents: 'none',
            zIndex: 1000,
          }}
        >
          {hoveredNode ? (
            <>
              <Typography variant="body2" sx={{ fontWeight: 500 }}>
                {hoveredNode.label}
              </Typography>
              <Typography variant="caption" color="text.secondary">
                {hoveredNode.type === 'contact' ? t('network.legend.contact') : t('network.legend.activity')}
              </Typography>
            </>
          ) : hoveredEdge ? (
            <>
              <Typography variant="body2" sx={{ fontWeight: 500 }}>
                {hoveredEdge.label}
              </Typography>
              <Typography variant="caption" color="text.secondary">
                {getEdgeTypeLabel(hoveredEdge.type)}
              </Typography>
            </>
          ) : null}
        </Box>
      )}
    </Box>
  );
}
