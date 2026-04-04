package server
import "net/http"
func(s *Server)dashboard(w http.ResponseWriter,r *http.Request){w.Header().Set("Content-Type","text/html");w.Write([]byte(dashHTML))}
const dashHTML=`<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Post</title>
<style>:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--mono:'JetBrains Mono',monospace;--serif:'Libre Baskerville',serif}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--serif);line-height:1.7}
.hdr{padding:1rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}.hdr h1{font-family:var(--mono);font-size:.9rem;letter-spacing:2px}
.main{padding:1.5rem;max-width:800px;margin:0 auto}
.post{border-bottom:1px solid var(--bg3);padding:1rem 0;cursor:pointer}
.post:hover{background:var(--bg2);margin:0 -1rem;padding:1rem}
.post-title{font-size:1.1rem;margin-bottom:.2rem}
.post-meta{font-family:var(--mono);font-size:.6rem;color:var(--cm);display:flex;gap:.8rem}
.post-excerpt{font-size:.85rem;color:var(--cd);margin-top:.3rem;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden}
.badge-draft{color:var(--gold);background:#d4a84322;border:1px solid #d4a84344;font-family:var(--mono);font-size:.5rem;padding:.1rem .3rem;text-transform:uppercase;letter-spacing:1px}
.badge-published{color:var(--green);background:#4a9e5c22;border:1px solid #4a9e5c44;font-family:var(--mono);font-size:.5rem;padding:.1rem .3rem;text-transform:uppercase;letter-spacing:1px}
.editor{display:none;max-width:800px;margin:0 auto;padding:1.5rem}.editor.open{display:block}
.editor input{width:100%;padding:.6rem .8rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--serif);font-size:1.2rem;margin-bottom:.5rem}
.editor textarea{width:100%;min-height:400px;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--serif);font-size:.95rem;padding:1rem;line-height:1.8;resize:vertical}
.editor-bar{display:flex;gap:.5rem;margin-bottom:.8rem;flex-wrap:wrap;align-items:center;font-family:var(--mono);font-size:.65rem}
.editor-bar input,.editor-bar select{padding:.3rem .5rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.65rem}
.btn{font-family:var(--mono);font-size:.6rem;padding:.3rem .7rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd)}.btn:hover{border-color:var(--leather);color:var(--cream)}
.btn-p{background:var(--rust);border-color:var(--rust);color:var(--bg)}
.btn-g{background:var(--green);border-color:var(--green);color:var(--bg)}
.empty{text-align:center;padding:3rem;color:var(--cm);font-style:italic;font-size:.85rem}
</style></head><body>
<div class="hdr"><h1>POST</h1><button class="btn btn-p" onclick="newPost()">+ New Post</button></div>
<div class="editor" id="editor">
<input id="e-title" placeholder="Post title...">
<div class="editor-bar"><label>Slug</label><input id="e-slug" style="width:150px"><label>Category</label><input id="e-cat" style="width:100px"><label>Author</label><input id="e-author" style="width:100px"><select id="e-status"><option value="draft">Draft</option><option value="published">Published</option></select></div>
<textarea id="e-body" placeholder="Write in Markdown..."></textarea>
<div style="display:flex;justify-content:space-between;margin-top:.5rem"><button class="btn" onclick="cancelEdit()">Cancel</button><div style="display:flex;gap:.4rem"><button class="btn" onclick="saveDraft()">Save Draft</button><button class="btn btn-g" onclick="publish()">Publish</button></div></div>
</div>
<div class="main" id="main"></div>
<script>
const A='/api';let articles=[],editId=null;
async function load(){const r=await fetch(A+'/articles').then(r=>r.json());articles=r.articles||[];render();}
function render(){if(!articles.length){document.getElementById('main').innerHTML='<div class="empty">No posts yet. Write your first one.</div>';return;}
const drafts=articles.filter(a=>a.status==='draft'),published=articles.filter(a=>a.status==='published');
let h='';
if(published.length){h+='<div style="font-family:var(--mono);font-size:.6rem;color:var(--green);text-transform:uppercase;letter-spacing:1px;margin-bottom:.5rem">Published ('+published.length+')</div>';published.forEach(a=>{h+=postCard(a)});}
if(drafts.length){h+='<div style="font-family:var(--mono);font-size:.6rem;color:var(--gold);margin:1rem 0 .5rem;text-transform:uppercase;letter-spacing:1px">Drafts ('+drafts.length+')</div>';drafts.forEach(a=>{h+=postCard(a)});}
document.getElementById('main').innerHTML=h;}
function postCard(a){let h='<div class="post" onclick="editPost(\''+a.id+'\')"><div style="display:flex;justify-content:space-between"><div class="post-title">'+esc(a.title||'Untitled')+'</div><span class="badge-'+a.status+'">'+a.status+'</span></div><div class="post-meta">';
if(a.author)h+='<span>'+esc(a.author)+'</span>';if(a.category)h+='<span>'+esc(a.category)+'</span>';
if(a.published_at)h+='<span>Published '+ft(a.published_at)+'</span>';else h+='<span>'+ft(a.created_at)+'</span>';
if(a.slug)h+='<span>/'+esc(a.slug)+'</span>';
h+='</div>';if(a.body)h+='<div class="post-excerpt">'+esc(a.body)+'</div>';h+='</div>';return h;}
function newPost(){editId=null;document.getElementById('e-title').value='';document.getElementById('e-body').value='';document.getElementById('e-slug').value='';document.getElementById('e-cat').value='';document.getElementById('e-author').value='';document.getElementById('e-status').value='draft';document.getElementById('editor').classList.add('open');document.getElementById('e-title').focus();}
function editPost(id){const a=articles.find(x=>x.id===id);if(!a)return;editId=id;document.getElementById('e-title').value=a.title||'';document.getElementById('e-body').value=a.body||'';document.getElementById('e-slug').value=a.slug||'';document.getElementById('e-cat').value=a.category||'';document.getElementById('e-author').value=a.author||'';document.getElementById('e-status').value=a.status;document.getElementById('editor').classList.add('open');}
function cancelEdit(){document.getElementById('editor').classList.remove('open');editId=null;}
async function saveDraft(){await savePost('draft');}
async function publish(){await savePost('published');}
async function savePost(status){const data={title:document.getElementById('e-title').value,body:document.getElementById('e-body').value,slug:document.getElementById('e-slug').value,category:document.getElementById('e-cat').value,author:document.getElementById('e-author').value,status};
if(editId){await fetch(A+'/articles/'+editId,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(data)});}
else{await fetch(A+'/articles',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(data)});}
cancelEdit();load();}
function ft(t){if(!t)return'';return new Date(t).toLocaleDateString();}
function esc(s){if(!s)return'';const d=document.createElement('div');d.textContent=s;return d.innerHTML;}
load();
</script></body></html>`
