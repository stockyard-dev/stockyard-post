package server
import "net/http"
func(s *Server)dashboard(w http.ResponseWriter,r *http.Request){w.Header().Set("Content-Type","text/html");w.Write([]byte(dashHTML))}
const dashHTML=`<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Post</title>
<style>:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--mono:'JetBrains Mono',monospace;--serif:'Libre Baskerville',serif}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--serif);line-height:1.7}
.hdr{padding:1rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}.hdr h1{font-family:var(--mono);font-size:.9rem;letter-spacing:2px}
.main{max-width:800px;margin:0 auto;padding:1.5rem}
.article{border-bottom:1px solid var(--bg3);padding:1.2rem 0;cursor:pointer}
.article:hover{background:var(--bg2);margin:0 -1rem;padding:1.2rem 1rem}
.article-title{font-size:1.1rem;margin-bottom:.2rem}
.article-meta{font-family:var(--mono);font-size:.6rem;color:var(--cm);display:flex;gap:.8rem}
.article-body{font-size:.85rem;color:var(--cd);margin-top:.3rem;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden}
.badge-draft{color:var(--cm)}.badge-published{color:var(--green)}
.editor{display:none;max-width:800px;margin:0 auto;padding:1.5rem}.editor.open{display:block}
.editor input{width:100%;padding:.6rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--serif);font-size:1.2rem;margin-bottom:.5rem}
.editor textarea{width:100%;min-height:400px;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--serif);font-size:.95rem;padding:1rem;line-height:1.8;resize:vertical}
.editor-bar{display:flex;gap:.5rem;margin-bottom:.8rem;align-items:center;font-family:var(--mono);font-size:.65rem}
.editor-bar input,.editor-bar select{padding:.3rem .5rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.btn{font-family:var(--mono);font-size:.65rem;padding:.3rem .7rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd)}.btn:hover{border-color:var(--leather);color:var(--cream)}
.btn-p{background:var(--rust);border-color:var(--rust);color:var(--bg)}
.btn-green{background:var(--green);border-color:var(--green);color:var(--bg)}
.empty{text-align:center;padding:3rem;color:var(--cm);font-style:italic}
</style></head><body>
<div class="hdr"><h1>POST</h1><button class="btn btn-p" onclick="newPost()">+ New Post</button></div>
<div class="editor" id="editor">
  <input id="e-title" placeholder="Post title">
  <div class="editor-bar">
    <label>Author</label><input id="e-author" style="width:120px">
    <label>Category</label><input id="e-cat" style="width:100px">
    <label>Slug</label><input id="e-slug" style="width:120px">
    <button class="btn" onclick="cancelEdit()">Cancel</button>
    <button class="btn" onclick="saveDraft()">Save Draft</button>
    <button class="btn btn-green" onclick="publish()">Publish</button>
  </div>
  <textarea id="e-body" placeholder="Write in Markdown..."></textarea>
</div>
<div class="main" id="main"></div>
<script>
const A='/api';let articles=[],editId=null;
async function load(){const r=await fetch(A+'/articles').then(r=>r.json());articles=r.articles||[];render();}
function render(){if(!articles.length){document.getElementById('main').innerHTML='<div class="empty">No posts yet. Write your first one.</div>';return;}
let h='';articles.forEach(a=>{
h+='<div class="article" onclick="editPost(\''+a.id+'\')"><div style="display:flex;justify-content:space-between"><div class="article-title">'+esc(a.title||'Untitled')+'</div><span class="badge-'+a.status+'" style="font-family:var(--mono);font-size:.55rem">'+a.status+'</span></div>';
h+='<div class="article-meta">';if(a.author)h+='<span>'+esc(a.author)+'</span>';if(a.published_at)h+='<span>'+ft(a.published_at)+'</span>';else h+='<span>'+ft(a.created_at)+'</span>';if(a.category)h+='<span>'+esc(a.category)+'</span>';if(a.slug)h+='<span>/'+esc(a.slug)+'</span>';h+='</div>';
h+='<div class="article-body">'+esc((a.body||'').substring(0,200))+'</div>';
h+='<div style="margin-top:.3rem"><button class="btn" onclick="event.stopPropagation();del(\''+a.id+'\')" style="font-size:.5rem;color:var(--cm)">Delete</button></div>';
h+='</div>';});document.getElementById('main').innerHTML=h;}
function newPost(){editId=null;document.getElementById('e-title').value='';document.getElementById('e-body').value='';document.getElementById('e-author').value='';document.getElementById('e-cat').value='';document.getElementById('e-slug').value='';document.getElementById('editor').classList.add('open');document.getElementById('e-title').focus();}
function editPost(id){const a=articles.find(x=>x.id===id);if(!a)return;editId=id;document.getElementById('e-title').value=a.title||'';document.getElementById('e-body').value=a.body||'';document.getElementById('e-author').value=a.author||'';document.getElementById('e-cat').value=a.category||'';document.getElementById('e-slug').value=a.slug||'';document.getElementById('editor').classList.add('open');}
function cancelEdit(){document.getElementById('editor').classList.remove('open');editId=null;}
async function saveDraft(){await savePost('draft');}
async function publish(){await savePost('published');}
async function savePost(status){const data={title:document.getElementById('e-title').value,body:document.getElementById('e-body').value,author:document.getElementById('e-author').value,category:document.getElementById('e-cat').value,slug:document.getElementById('e-slug').value,status};
if(editId){await fetch(A+'/articles/'+editId,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(data)});}
else{await fetch(A+'/articles',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(data)});}
cancelEdit();load();}
async function del(id){if(confirm('Delete?')){await fetch(A+'/articles/'+id,{method:'DELETE'});load();}}
function ft(t){if(!t)return'';return new Date(t).toLocaleDateString();}
function esc(s){if(!s)return'';const d=document.createElement('div');d.textContent=s;return d.innerHTML;}
load();
</script></body></html>`
