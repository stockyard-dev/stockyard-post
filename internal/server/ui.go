package server

import "net/http"

const uiHTML = `<!DOCTYPE html><html lang="en"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Post — Stockyard</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:ital,wght@0,400;0,700;1,400&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
<style>:root{
  --bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;
  --rust:#c45d2c;--rust-light:#e8753a;--rust-dark:#8b3d1a;
  --leather:#a0845c;--leather-light:#c4a87a;
  --cream:#f0e6d3;--cream-dim:#bfb5a3;--cream-muted:#7a7060;
  --gold:#d4a843;--green:#5ba86e;--red:#c0392b;
  --font-serif:'Libre Baskerville',Georgia,serif;
  --font-mono:'JetBrains Mono',monospace;
}
*{margin:0;padding:0;box-sizing:border-box}
body{background:var(--bg);color:var(--cream);font-family:var(--font-serif);min-height:100vh}
a{color:var(--rust-light);text-decoration:none}a:hover{color:var(--gold)}
.hdr{background:var(--bg2);border-bottom:2px solid var(--rust-dark);padding:.9rem 1.8rem;display:flex;align-items:center;justify-content:space-between}
.hdr-left{display:flex;align-items:center;gap:1rem}
.hdr-brand{font-family:var(--font-mono);font-size:.75rem;color:var(--leather);letter-spacing:3px;text-transform:uppercase}
.hdr-title{font-family:var(--font-mono);font-size:1.1rem;color:var(--cream);letter-spacing:1px}
.badge{font-family:var(--font-mono);font-size:.6rem;padding:.2rem .6rem;letter-spacing:1px;text-transform:uppercase;border:1px solid}
.badge-free{color:var(--green);border-color:var(--green)}
.main{max-width:1000px;margin:0 auto;padding:2rem 1.5rem}
.cards{display:grid;grid-template-columns:repeat(auto-fit,minmax(160px,1fr));gap:1rem;margin-bottom:2rem}
.card{background:var(--bg2);border:1px solid var(--bg3);padding:1.2rem 1.5rem}
.card-val{font-family:var(--font-mono);font-size:1.8rem;font-weight:700;color:var(--cream);display:block}
.card-lbl{font-family:var(--font-mono);font-size:.62rem;letter-spacing:2px;text-transform:uppercase;color:var(--leather);margin-top:.3rem}
.section{margin-bottom:2.5rem}
.section-title{font-family:var(--font-mono);font-size:.68rem;letter-spacing:3px;text-transform:uppercase;color:var(--rust-light);margin-bottom:.8rem;padding-bottom:.5rem;border-bottom:1px solid var(--bg3)}
table{width:100%;border-collapse:collapse;font-family:var(--font-mono);font-size:.75rem}
th{background:var(--bg3);padding:.5rem .8rem;text-align:left;color:var(--leather-light);font-weight:400;letter-spacing:1px;font-size:.62rem;text-transform:uppercase}
td{padding:.5rem .8rem;border-bottom:1px solid var(--bg3);color:var(--cream-dim);vertical-align:top;word-break:break-all}
tr:hover td{background:var(--bg2)}
.empty{color:var(--cream-muted);text-align:center;padding:2rem;font-style:italic}
.btn{font-family:var(--font-mono);font-size:.7rem;padding:.3rem .8rem;border:1px solid var(--leather);background:transparent;color:var(--cream);cursor:pointer;transition:all .2s}
.btn:hover{border-color:var(--rust-light);color:var(--rust-light)}
.btn-rust{border-color:var(--rust);color:var(--rust-light)}.btn-rust:hover{background:var(--rust);color:var(--cream)}
.btn-sm{font-size:.62rem;padding:.2rem .5rem}
.lbl{font-family:var(--font-mono);font-size:.62rem;letter-spacing:1px;text-transform:uppercase;color:var(--leather)}
input{font-family:var(--font-mono);font-size:.78rem;background:var(--bg3);border:1px solid var(--bg3);color:var(--cream);padding:.4rem .7rem;outline:none}
input:focus{border-color:var(--leather)}
.row{display:flex;gap:.8rem;align-items:flex-end;flex-wrap:wrap;margin-bottom:1rem}
.field{display:flex;flex-direction:column;gap:.3rem}
.tabs{display:flex;gap:0;margin-bottom:1.5rem;border-bottom:1px solid var(--bg3)}
.tab{font-family:var(--font-mono);font-size:.72rem;padding:.6rem 1.2rem;color:var(--cream-muted);cursor:pointer;border-bottom:2px solid transparent;letter-spacing:1px;text-transform:uppercase}
.tab:hover{color:var(--cream-dim)}.tab.active{color:var(--rust-light);border-bottom-color:var(--rust-light)}
.tab-content{display:none}.tab-content.active{display:block}
pre{background:var(--bg3);padding:.8rem 1rem;font-family:var(--font-mono);font-size:.72rem;color:var(--cream-dim);overflow-x:auto}
</style></head><body>
<div class="hdr">
  <div class="hdr-left">
    <svg viewBox="0 0 64 64" width="22" height="22" fill="none"><rect x="8" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="28" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="48" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="8" y="27" width="48" height="7" rx="2.5" fill="#c4a87a"/></svg>
    <span class="hdr-brand">Stockyard</span>
    <span class="hdr-title">Post</span>
  </div>
  <div style="display:flex;gap:.8rem;align-items:center">
    <span class="badge badge-free">Free</span>
    <a href="/api/status" class="lbl" style="color:var(--leather)">API</a>
  </div>
</div>
<div class="main">

<div class="cards">
  <div class="card"><span class="card-val" id="s-forms">—</span><span class="card-lbl">Forms</span></div>
  <div class="card"><span class="card-val" id="s-subs">—</span><span class="card-lbl">Submissions</span></div>
</div>

<div class="tabs">
  <div class="tab active" onclick="switchTab('forms')">Forms</div>
  <div class="tab" onclick="switchTab('create')">Create</div>
  <div class="tab" onclick="switchTab('submissions')">Submissions</div>
  <div class="tab" onclick="switchTab('usage')">Usage</div>
</div>

<div id="tab-forms" class="tab-content active">
  <div class="section">
    <div class="section-title">Forms</div>
    <table><thead><tr><th>Name</th><th>Endpoint</th><th>Submissions</th><th>Created</th><th></th></tr></thead>
    <tbody id="forms-body"></tbody></table>
  </div>
</div>

<div id="tab-create" class="tab-content">
  <div class="section">
    <div class="section-title">Create Form</div>
    <div class="row">
      <div class="field"><span class="lbl">Name</span><input id="c-name" placeholder="Contact" style="width:160px"></div>
      <div class="field"><span class="lbl">Redirect URL (optional)</span><input id="c-redirect" placeholder="https://mysite.com/thanks" style="width:260px"></div>
      <button class="btn btn-rust" onclick="createForm()">Create</button>
    </div>
    <div id="c-result" style="margin-top:.5rem"></div>
  </div>
</div>

<div id="tab-submissions" class="tab-content">
  <div class="section">
    <div class="section-title">Recent Submissions</div>
    <div id="subs-list"></div>
  </div>
</div>

<div id="tab-usage" class="tab-content">
  <div class="section">
    <div class="section-title">Quick Start</div>
    <pre>
&lt;!-- Point any HTML form at your Post endpoint --&gt;
&lt;form method="POST" action="http://localhost:8830/f/{form_id}"&gt;
  &lt;input name="email" type="email" required&gt;
  &lt;input name="message" type="text"&gt;
  &lt;button type="submit"&gt;Send&lt;/button&gt;
&lt;/form&gt;

# Or submit via API
curl -X POST http://localhost:8830/f/{form_id} \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"email":"user@example.com","message":"Hello!"}'

# List submissions
curl http://localhost:8830/api/forms/{form_id}/submissions
    </pre>
  </div>
</div>

</div>
<script>
let forms=[];

function switchTab(n){
  document.querySelectorAll('.tab').forEach(t=>t.classList.toggle('active',t.textContent.toLowerCase()===n));
  document.querySelectorAll('.tab-content').forEach(t=>t.classList.toggle('active',t.id==='tab-'+n));
  if(n==='submissions')loadSubs();
}

async function refresh(){
  try{
    const r=await fetch('/api/status');const s=await r.json();
    document.getElementById('s-forms').textContent=s.forms||0;
    document.getElementById('s-subs').textContent=fmt(s.submissions||0);
  }catch(e){}
  try{
    const r=await fetch('/api/forms');const d=await r.json();
    forms=d.forms||[];
    const tb=document.getElementById('forms-body');
    if(!forms.length){tb.innerHTML='<tr><td colspan="5" class="empty">No forms yet.</td></tr>';return;}
    tb.innerHTML=forms.map(f=>
      '<tr><td style="color:var(--cream);font-weight:600">'+esc(f.name)+'</td>'+
      '<td style="font-size:.65rem"><code>/f/'+f.id+'</code></td>'+
      '<td>'+f.submission_count+'</td>'+
      '<td style="font-size:.65rem;color:var(--cream-muted)">'+timeAgo(f.created_at)+'</td>'+
      '<td><button class="btn btn-sm" onclick="deleteForm(\''+f.id+'\')">Delete</button></td></tr>'
    ).join('');
  }catch(e){}
}

async function createForm(){
  const name=document.getElementById('c-name').value.trim();
  const redirect=document.getElementById('c-redirect').value.trim();
  if(!name){document.getElementById('c-result').innerHTML='<span style="color:var(--red)">Name required</span>';return;}
  const r=await fetch('/api/forms',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name,redirect_url:redirect})});
  const d=await r.json();
  if(r.ok){
    document.getElementById('c-result').innerHTML='<span style="color:var(--green)">Created! Endpoint: <code>'+d.submit_url+'</code></span>';
    document.getElementById('c-name').value='';document.getElementById('c-redirect').value='';
    refresh();
  }else{document.getElementById('c-result').innerHTML='<span style="color:var(--red)">'+esc(d.error)+'</span>';}
}

async function deleteForm(id){
  if(!confirm('Delete form and all submissions?'))return;
  await fetch('/api/forms/'+id,{method:'DELETE'});
  refresh();
}

async function loadSubs(){
  let html='';
  for(const f of forms){
    const r=await fetch('/api/forms/'+f.id+'/submissions?limit=10');
    const d=await r.json();
    const subs=d.submissions||[];
    html+='<div class="section-title" style="margin-top:1rem">'+esc(f.name)+'</div>';
    if(!subs.length){html+='<div class="empty">No submissions</div>';continue;}
    html+='<table><thead><tr><th>Data</th><th>IP</th><th>Time</th></tr></thead><tbody>';
    html+=subs.map(s=>'<tr><td style="font-size:.65rem">'+esc(s.data)+'</td><td style="font-size:.65rem">'+esc(s.source_ip)+'</td><td style="font-size:.65rem">'+timeAgo(s.created_at)+'</td></tr>').join('');
    html+='</tbody></table>';
  }
  document.getElementById('subs-list').innerHTML=html||'<div class="empty">No forms yet</div>';
}

function fmt(n){if(n>=1e6)return(n/1e6).toFixed(1)+'M';if(n>=1e3)return(n/1e3).toFixed(1)+'K';return n;}
function esc(s){const d=document.createElement('div');d.textContent=s||'';return d.innerHTML;}
function timeAgo(s){if(!s)return'—';const d=new Date(s);const diff=Date.now()-d.getTime();if(diff<60000)return'now';if(diff<3600000)return Math.floor(diff/60000)+'m';if(diff<86400000)return Math.floor(diff/3600000)+'h';return Math.floor(diff/86400000)+'d';}

refresh();
setInterval(refresh,8000);
</script></body></html>`

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(uiHTML))
}
