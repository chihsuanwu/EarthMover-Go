const D3 = {};

D3.requestTree = async function() {
  try {
    const response = await post(null, '/visualize');
    if (response) D3.drawTree(response);
  } catch (e) {
    // visualization request failed, ignore
  }
};

D3.removeTree = function() {
  d3.select('#tv-svg').select('g').remove();
};

D3.drawTree = function(treeData) {
  const root = d3.hierarchy(treeData, d => d.ch)
    .sort((a, b) => b.data.tc - a.data.tc);

  const margin = { left: 60, top: 40, right: 60, bottom: 40 };

  const svg = d3.select('#tv-svg');

  const g = svg.select('g').empty()
    ? svg.append('g').attr('transform', `translate(${margin.left},${margin.top})`)
    : svg.select('g');

  const duration = 600;

  let verticalSpacing, horizontalSpacing;

  update();

  svg.transition().duration(duration)
    .attr('width', root.height * horizontalSpacing + margin.left + margin.right)
    .attr('height', (root.leaves().length - 1) * verticalSpacing + margin.top + margin.bottom);

  function update() {
    const t = d3.transition().duration(duration);

    const maxBranch = parseInt(document.getElementById('tv-branch-factor').value, 10);
    const minCount = parseInt(document.getElementById('tv-count-threshold').value, 10);

    hideChildren(root, maxBranch, minCount);

    function hideChildren(node, branch, count) {
      if (!node.children) return;
      let reserved = 0;
      for (reserved = 0; reserved < branch && reserved < node.children.length; ++reserved) {
        if (node.children[reserved].data.tc < count) break;
      }
      if (reserved < 1) {
        node.children = null;
        return;
      }
      node.children = node.children.slice(0, reserved);
      node.children.forEach(d => hideChildren(d, branch, count));
    }

    verticalSpacing = 300 / root.leaves().length + 10;

    let deepest = 0;

    root.leaves().forEach((d, i) => {
      d.x = i * verticalSpacing;
      if (d.depth > deepest) deepest = d.depth;
    });

    horizontalSpacing = 240 / (deepest + 1) + 60;

    const nodes = root.descendants();
    const links = nodes.slice(1);

    nodes.forEach(d => {
      d.x = d.leaves()[0].x;
      d.y = d.depth * horizontalSpacing;
    });

    // ---- Nodes ----

    const node = g.selectAll('g.node')
      .data(nodes, function(d) {
        if (!d.id) {
          d.id = 'id' + d.data.i;
          let currentNode = d;
          while (currentNode = currentNode.parent) {
            d.id += currentNode.data.i;
          }
        }
        return d.id;
      });

    const nodeEnter = node.enter().append('g')
      .attr('id', d => d.id)
      .attr('class', 'node')
      .attr('transform', function(d) {
        let currentNode = d;
        while (currentNode = currentNode.parent) {
          if (d.previousPos = d3.select('#' + currentNode.id).property('previousPos')) break;
        }
        if (!d.previousPos) d.previousPos = { x: 0, y: 0 };
        return `translate(${d.previousPos.y},${d.previousPos.x})`;
      });

    nodeEnter.append('circle')
      .attr('r', 0)
      .style('fill', d => d.data.wt ? '#444' : '#ffe')
      .style('stroke', d => gradient(d.data.wr));

    nodeEnter.append('text')
      .attr('x', 12)
      .attr('y', 0)
      .attr('text-anchor', 'start')
      .style('font-size', 0)
      .text(d =>
        d.data.i === 225
          ? '(pass)'
          : `(${String.fromCharCode(65 + d.data.i % 15)}${Math.floor(d.data.i / 15) + 1})`
      )
      .style('fill-opacity', 0);

    nodeEnter.append('line')
      .filter(d => d.data.wol !== 0)
      .attr('x1', 0).attr('x2', 0)
      .attr('y1', 0).attr('y2', 0);

    // UPDATE
    nodeEnter.merge(node).on('click', click);
    nodeEnter.merge(node).on('mouseover', mouseOver);
    nodeEnter.merge(node).on('mouseout', mouseOut);

    const nodeUpdate = nodeEnter.merge(node).transition(t)
      .attr('transform', d => `translate(${d.y},${d.x})`);

    nodeUpdate.each(function(d) {
      d3.select(this).property('previousPos', { x: d.x, y: d.y });

      const w = width(d.data.tc);
      const r = w * 0.75 + 'px';
      const cw = w * 0.25 + 'px';
      const c = gradient(d.data.wr);
      let l = 0, nl = 0;
      const lw = w * 0.2 + 'px';
      let s = null;
      const f = w + 'px';
      const dy = w * 0.33 + 'px';

      if (d.data.wol !== 0) {
        l = w * 0.44 + 'px';
        nl = -w * 0.44 + 'px';
        s = d.data.wol === 1 ? gradient(1) : gradient(0);
      }

      const upd = d3.select(this).transition(t);

      upd.select('circle')
        .attr('r', r)
        .style('stroke', c)
        .style('stroke-width', cw);

      upd.select('line')
        .attr('x1', nl).attr('x2', l)
        .attr('y1', nl).attr('y2', l)
        .style('stroke-width', lw)
        .style('stroke', s);

      upd.select('text')
        .attr('dy', dy)
        .style('font-size', f)
        .style('fill-opacity', 1);
    });

    // EXIT
    const nodeExit = node.exit().transition(t)
      .attr('transform', function(d) {
        let currentNode = d;
        while (currentNode = currentNode.parent) {
          if (!currentNode.exitPos) {
            currentNode = d3.select('#' + currentNode.id).datum();
            d.exitPos = { x: currentNode.x, y: currentNode.y };
            break;
          }
        }
        if (!d.exitPos) d.exitPos = { x: 0, y: 0 };
        return `translate(${d.exitPos.y},${d.exitPos.x})`;
      })
      .remove();

    nodeExit.select('circle').attr('r', 0);
    nodeExit.select('text').style('fill-opacity', 0);
    nodeExit.select('line')
      .attr('x1', 0).attr('x2', 0)
      .attr('y1', 0).attr('y2', 0);

    // ---- Links ----

    const link = g.selectAll('path.link')
      .data(links, d => d.id);

    const linkEnter = link.enter().insert('path', 'g')
      .attr('class', 'link')
      .attr('d', d => diagonal(d.previousPos, d.previousPos));

    linkEnter.merge(link).transition(t)
      .attr('d', d => diagonal(d, d.parent))
      .style('stroke-width', d => width(d.data.tc) + 'px')
      .style('stroke', d => gradient(d.data.wr));

    link.exit().transition(t)
      .attr('d', function(d) {
        const c = diagonal(d.exitPos, d.exitPos);
        d.exitPos = null;
        return c;
      })
      .remove();

    function diagonal(s, d) {
      return `M ${s.y} ${s.x} C ${(s.y + d.y) / 2} ${s.x}, ${(s.y + d.y) / 2} ${d.x}, ${d.y} ${d.x}`;
    }

    // D3 v4: event handlers receive datum as first arg
    function click(d) {
      if (d.children) {
        d.hiddenChildren = d.children;
        d.children = null;
      } else {
        d.children = d.hiddenChildren;
        d.hiddenChildren = null;
      }
      update();
    }

    function mouseOver(d) {
      updateInf(d);
      const evolve = [];
      while (d.parent) {
        evolve.unshift({
          row: Math.floor(d.data.i / 15),
          col: Math.floor(d.data.i % 15),
          color: d.data.wt ? 'black' : 'white'
        });
        d = d.parent;
      }
      if (evolve.length) board.drawEvolve(evolve);
    }

    mouseOut();

    function mouseOut() {
      updateInf(root.children ? root.children[0] : root);
      board.drawAll();
    }

    function updateInf(d) {
      d3.select('#tv-inf-position')
        .text(String.fromCharCode(65 + d.data.i % 15) + (Math.floor(d.data.i / 15) + 1));
      d3.select('#tv-inf-count').text('count: ' + d.data.tc);
      const winRate = d.data.wr;
      d3.select('#tv-inf-win-rate-pointer')
        .style('left', (38 + 120 * winRate) + 'px')
        .style('background', gradient(winRate));
      const pct = Math.floor(winRate * 100);
      d3.select('#tv-inf-win-rate-white').text(100 - pct);
      d3.select('#tv-inf-win-rate-black').text(pct);
    }

    function gradient(level) {
      let h;
      if (level < 19 / 35) {
        h = (Math.asin(level * 3.5 - 0.9) * 180 / Math.PI + 90) * (2 / 3);
      } else {
        h = 120 + (Math.asin(level * 3.5 - 2.9) * 180 / Math.PI + 90) * (2 / 3);
      }
      return `hsl(${h},80%,50%)`;
    }

    function width(count) {
      return Math.pow(count, 0.24) + 6;
    }
  }
};
