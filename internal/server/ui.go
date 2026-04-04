package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Post</title>
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:ital,wght@0,400;0,700;1,400&family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet">
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c94444;--mono:'JetBrains Mono',monospace;--serif:'Libre Baskerville',serif}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--serif);line-height:1.7}
.hdr{padding:1rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}.hdr h1{font-family:var(--mono);font-size:.9rem;letter-spacing:2px}.hdr h1 span{color:var(--rust)}
.main{max-width:900px;margin:0 auto;padding:1.5rem}
.stats{display:grid;grid-template-columns:repeat(3,1fr);gap:.5rem;margin-bottom:1rem}
.st{background:var(--bg2);border:1px solid var(--bg3);padding:.6rem;text-align:center;font-family:var(--mono);cursor:pointer;transition:border-color .2s}
.st:hover,.st.active{border-color:var(--rust)}.st.active .st-v{color:var(--rust)}
.st-v{font-size:1.3rem;font-weight:700}.st-l{font-size:.5rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-top:.15rem}
.toolbar{display:flex;gap:.5rem;margin-bottom:1rem;align-items:center}
.search{flex:1;padding:.4rem .6rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.search:focus{outline:none;border-color:var(--leather)}
.article{background:var(--bg2);border:1px solid var(--bg3);padding:1rem;margin-bottom:.5rem;cursor:pointer;transition:border-color .2s}
.article:hover{border-color:var(--leather)}
.article-top{display:flex;justify-content:space-between;align-items:flex-start;gap:.5rem}
.article-title{font-size:1rem;margin-bottom:.15rem}
.article-meta{font-family:var(--mono);font-size:.55rem;color:var(--cm);display:flex;gap:.6rem;flex-wrap:wrap;margin-top:.3rem}
.article-excerpt{font-size:.8rem;color:var(--cd);margin-top:.4rem;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden}
.badge{font-family:var(--mono);font-size:.5rem;padding:.15rem .4rem;text-transform:uppercase;letter-spacing:1px;border:1px solid;flex-shrink:0}
.badge.draft{border-color:var(--cm);color:var(--cm)}.badge.published{border-color:var(--green);color:var(--green)}
.cat-badge{font-family:var(--mono);font-size:.5rem;padding:.1rem .35rem;background:var(--bg3);color:var(--cm)}
.btn{font-family:var(--mono);font-size:.6rem;padding:.3rem .6rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd);transition:all .2s}
.btn:hover{border-color:var(--leather);color:var(--cream)}.btn-p{background:var(--rust);border-color:var(--rust);color:#fff}.btn-p:hover{background:#d4682f}
.btn-green{background:var(--green);border-color:var(--green);color:#fff}
.btn-sm{font-size:.55rem;padding:.2rem .4rem}
.empty{text-align:center;padding:3rem;color:var(--cm);font-style:italic;font-size:.85rem}

.editor-overlay{display:none;position:fixed;inset:0;background:var(--bg);z-index:100;overflow-y:auto}.editor-overlay.open{display:block}
.editor{max-width:800px;margin:0 auto;padding:1.5rem}
.editor-hdr{display:flex;justify-content:space-between;align-items:center;margin-bottom:1rem;padding-bottom:.8rem;border-bottom:1px solid var(--bg3)}
.editor-hdr h2{font-family:var(--mono);font-size:.8rem;color:var(--rust);letter-spacing:1px}
.editor-actions{display:flex;gap:.4rem}
.e-title{width:100%;padding:.6rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--serif);font-size:1.3rem;margin-bottom:.8rem}
.e-title:focus{outline:none;border-color:var(--leather)}
.e-fields{display:grid;grid-template-columns:1fr 1fr 1fr;gap:.5rem;margin-bottom:.8rem}
.fr{margin-bottom:0}.fr label{display:block;font-family:var(--mono);font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-bottom:.2rem}
.fr input,.fr select{width:100%;padding:.35rem .5rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.fr input:focus,.fr select:focus{outline:none;border-color:var(--leather)}
.e-body{width:100%;min-height:450px;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.85rem;padding:1rem;line-height:1.8;resize:vertical;tab-size:2}
.e-body:focus{outline:none;border-color:var(--leather)}
.e-footer{display:flex;justify-content:space-between;align-items:center;margin-top:.5rem;font-family:var(--mono);font-size:.55rem;color:var(--cm)}
@media(max-width:600px){.stats{grid-template-columns:repeat(3,1fr)}.e-fields{grid-template-columns:1fr}.toolbar{flex-direction:column}.search{width:100%}}
</style></head><body>

<div class="hdr"><h1><span>&#9670;</span> POST</h1><button class="btn btn-p" onclick="newPost()">+ New Post</button></div>
<div class="main">
<div class="stats" id="stats"></div>
<div class="toolbar"><input class="search" id="search" type="text" placeholder="Search posts..." oninput="render()"></div>
<div id="list"></div>
</div>

<div class="editor-overlay" id="editorOv">
<div class="editor">
<div class="editor-hdr">
<h2 id="editorTitle">NEW POST</h2>
<div class="editor-actions">
<button class="btn" onclick="cancelEdit()">&#10005; Close</button>
<button class="btn" onclick="saveDraft()">Save Draft</button>
<button class="btn btn-green" onclick="publish()">&#10003; Publish</button>
</div>
</div>
<input class="e-title" id="e-title" placeholder="Post title..." oninput="autoSlug()">
<div class="e-fields">
<div class="fr"><label>Author</label><input id="e-author"></div>
<div class="fr"><label>Category</label><input id="e-cat" placeholder="e.g. engineering"></div>
<div class="fr"><label>Slug</label><input id="e-slug" placeholder="auto-generated"></div>
</div>
<textarea class="e-body" id="e-body" placeholder="Write your post..." oninput="updateWordCount()"></textarea>
<div class="e-footer"><span id="wordcount">0 words</span><span id="autosave"></span></div>
</div>
</div>

<script>
var A='/api',articles=[],filter='all',editId=null;

async function load(){var r=await fetch(A+'/articles').then(function(r){return r.json()});articles=r.articles||[];renderStats();render();}

function renderStats(){
var total=articles.length;
var drafts=articles.filter(function(a){return a.status==='draft'}).length;
var pub=articles.filter(function(a){return a.status==='published'}).length;
document.getElementById('stats').innerHTML=[
{l:'Total',v:total,f:'all'},
{l:'Drafts',v:drafts,f:'draft'},
{l:'Published',v:pub,f:'published'}
].map(function(x){return '<div class="st'+(filter===x.f?' active':'')+'" onclick="setFilter(\''+x.f+'\')"><div class="st-v">'+x.v+'</div><div class="st-l">'+x.l+'</div></div>'}).join('');
}

function setFilter(f){filter=f;renderStats();render();}

function render(){
var q=(document.getElementById('search').value||'').toLowerCase();
var f=articles;
if(filter!=='all')f=f.filter(function(a){return a.status===filter});
if(q)f=f.filter(function(a){return(a.title||'').toLowerCase().includes(q)||(a.body||'').toLowerCase().includes(q)||(a.category||'').toLowerCase().includes(q)});
if(!f.length){document.getElementById('list').innerHTML='<div class="empty">'+(articles.length?'No matching posts.':'No posts yet. Write your first one.')+'</div>';return;}
var h='';f.forEach(function(a){
h+='<div class="article" onclick="editPost(\''+a.id+'\')">';
h+='<div class="article-top"><div class="article-title">'+esc(a.title||'Untitled')+'</div>';
h+='<span class="badge '+a.status+'">'+a.status+'</span></div>';
h+='<div class="article-meta">';
if(a.author)h+='<span>'+esc(a.author)+'</span>';
h+='<span>'+ft(a.status==='published'&&a.published_at?a.published_at:a.created_at)+'</span>';
if(a.category)h+='<span class="cat-badge">'+esc(a.category)+'</span>';
if(a.slug)h+='<span>/'+esc(a.slug)+'</span>';
var wc=wordCount(a.body);if(wc)h+='<span>'+wc+' words</span>';
h+='</div>';
if(a.body){h+='<div class="article-excerpt">'+esc((a.body||'').substring(0,250))+'</div>';}
h+='</div>';
});
document.getElementById('list').innerHTML=h;
}

function wordCount(s){if(!s)return 0;return s.trim().split(/\s+/).filter(function(w){return w}).length;}

function newPost(){editId=null;document.getElementById('e-title').value='';document.getElementById('e-body').value='';document.getElementById('e-author').value='';document.getElementById('e-cat').value='';document.getElementById('e-slug').value='';document.getElementById('editorTitle').textContent='NEW POST';document.getElementById('editorOv').classList.add('open');document.getElementById('e-title').focus();updateWordCount();}

function editPost(id){
var a=null;for(var j=0;j<articles.length;j++){if(articles[j].id===id){a=articles[j];break;}}
if(!a)return;editId=id;
document.getElementById('e-title').value=a.title||'';
document.getElementById('e-body').value=a.body||'';
document.getElementById('e-author').value=a.author||'';
document.getElementById('e-cat').value=a.category||'';
document.getElementById('e-slug').value=a.slug||'';
document.getElementById('editorTitle').textContent='EDIT POST';
document.getElementById('editorOv').classList.add('open');
updateWordCount();
}

function cancelEdit(){document.getElementById('editorOv').classList.remove('open');editId=null;}

function autoSlug(){
var title=document.getElementById('e-title').value;
var slug=document.getElementById('e-slug');
if(!editId||!slug.value){
slug.value=title.toLowerCase().replace(/[^a-z0-9]+/g,'-').replace(/^-|-$/g,'');
}
}

function updateWordCount(){
var body=document.getElementById('e-body').value;
document.getElementById('wordcount').textContent=wordCount(body)+' words';
}

async function saveDraft(){await savePost('draft');}
async function publish(){await savePost('published');}

async function savePost(status){
var title=document.getElementById('e-title').value.trim();
if(!title){alert('Title is required');return;}
var data={title:title,body:document.getElementById('e-body').value,author:document.getElementById('e-author').value.trim(),category:document.getElementById('e-cat').value.trim(),slug:document.getElementById('e-slug').value.trim(),status:status};
if(status==='published')data.published_at=new Date().toISOString();
if(editId){await fetch(A+'/articles/'+editId,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(data)});}
else{await fetch(A+'/articles',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(data)});}
cancelEdit();load();
}

async function del(id){if(!confirm('Delete this post?'))return;await fetch(A+'/articles/'+id,{method:'DELETE'});load();}

function ft(t){if(!t)return'';try{return new Date(t).toLocaleDateString('en-US',{month:'short',day:'numeric',year:'numeric'})}catch(e){return t;}}
function esc(s){if(!s)return'';var d=document.createElement('div');d.textContent=s;return d.innerHTML;}

document.addEventListener('keydown',function(e){
if(e.key==='Escape')cancelEdit();
if((e.ctrlKey||e.metaKey)&&e.key==='s'&&document.getElementById('editorOv').classList.contains('open')){e.preventDefault();saveDraft();}
});
load();
</script></body></html>`
