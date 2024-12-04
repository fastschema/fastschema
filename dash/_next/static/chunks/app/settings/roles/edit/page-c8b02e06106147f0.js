(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[757],{73421:function(e,t,r){Promise.resolve().then(r.bind(r,44131))},44131:function(e,t,r){"use strict";r.r(t),r.d(t,{default:function(){return K}});var n=r(57437),s=r(76540),l=r(34560),a=r(2265),i=r(47907),o=r(75006),d=r(15e3),c=r(21270),u=r(82670),m=r(33108),f=r(30248);let x=f.z.object({name:f.z.string().trim().min(1,{message:"Role name is required"}),description:f.z.string().optional(),root:f.z.boolean().optional(),rule:f.z.string().optional(),permissions:f.z.array(f.z.object({resource:f.z.string(),value:f.z.string()})).optional(),$add:f.z.record(f.z.string(),f.z.array(m.gk)).optional(),$clear:f.z.record(f.z.string(),f.z.array(m.gk)).optional()}),h={name:"",description:"",root:!1,rule:"",permissions:[],$add:{},$clear:{}};var p=r(80244),g=r(82628),v=r(22782),j=r(3549),b=r(86468),N=r(12647),y=r(57654),w=r(27453),$=r(90688),z=r(26739),k=r(82012),C=r(95320),R=r(62177),S=r(22169);let _=C.fC,P=a.forwardRef((e,t)=>{let{className:r,...s}=e;return(0,n.jsx)(C.ck,{ref:t,className:(0,S.cn)("border-b",r),...s})});P.displayName="AccordionItem";let E=a.forwardRef((e,t)=>{let{className:r,children:s,...l}=e;return(0,n.jsx)(C.h4,{className:"flex",children:(0,n.jsxs)(C.xz,{ref:t,className:(0,S.cn)("flex flex-1 items-center justify-between py-4 text-sm font-medium transition-all hover:underline [&[data-state=open]>svg]:rotate-180",r),...l,children:[s,(0,n.jsx)(R.v4q,{className:"h-4 w-4 shrink-0 text-muted-foreground transition-transform duration-200"})]})})});E.displayName=C.xz.displayName;let O=a.forwardRef((e,t)=>{let{className:r,children:s,...l}=e;return(0,n.jsx)(C.VY,{ref:t,className:"overflow-hidden text-sm data-[state=closed]:animate-accordion-up data-[state=open]:animate-accordion-down",...l,children:(0,n.jsx)("div",{className:(0,S.cn)("pb-4 pt-0",r),children:s})})});O.displayName=C.VY.displayName;var V=r(62677),F=r(97081);r(33277);var I=r(15474),J=r(50489),A=r(29910);r(52235);var U=r(79134),Z=r(95453),q=r(51387),D=r(37033);let X=e=>{let{value:t,onChange:r,className:s,placeholder:l,height:i}=e,o=(0,a.useCallback)((e,t)=>{null==r||r(e)},[r]);return(0,n.jsx)(q.ZP,{className:(0,S.cn)("flex min-h-[60px] w-full rounded-md border border-input bg-transparent text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50",s),value:t,height:null!=i?i:"350px",width:"100%",placeholder:l,extensions:[(0,D.eJ)()],onChange:o})};var G=r(79984);let Y=e=>e.reduce((e,t)=>t.resources?[...e,...Y(t.resources)]:[...e,t],[]),T=e=>{var t,r;let s=(0,G.MG)(),{editingRole:l,onPermissionsUpdated:i,onUsersUpdated:o}=e,[d,c]=(0,a.useState)(!1),[u,m]=(0,a.useState)(null!==(t=null==l?void 0:l.permissions)&&void 0!==t?t:[]),f=s.resources,x=e=>{m(e),null==i||i(e)};return(0,n.jsxs)(_,{type:"multiple",className:"bg-slate-50 rounded-lg border",defaultValue:["permissions"],children:[(0,n.jsxs)(P,{value:"permissions",className:"border-0",children:[(0,n.jsx)(E,{className:"hover:no-underline px-4",children:"Permissions"}),(0,n.jsx)(O,{className:"pb-2",children:l?(0,n.jsxs)("div",{className:"px-4 pb-4 bg-background border-t",children:[(0,n.jsx)("div",{className:"flex flex-wrap gap-2 pt-4",children:(0,n.jsx)(Q,{resources:f,open:d,setOpen:c,selectedPermissions:u,updateSelectedPermissions:x})}),(0,n.jsx)("div",{className:"flex flex-col flex-wrap gap-2 mt-4",children:u.toSorted((e,t)=>e.resource.localeCompare(t.resource)).map(e=>(0,n.jsx)(L,{resources:Y(null!=f?f:[]),permission:e,selectedPermissions:u,updateSelectedPermissions:x},e.resource))})]}):(0,n.jsx)("p",{className:"px-4 py-4",children:"Please save the role before updating permissions."})})]}),(0,n.jsxs)(P,{value:"users",className:"border-t border-b-0",children:[(0,n.jsx)(E,{className:"hover:no-underline px-4",children:"Users"}),(0,n.jsx)(O,{className:"pb-2",children:(0,n.jsx)("div",{className:"px-4 pb-4 bg-background border-t",children:(0,n.jsx)("div",{className:"flex flex-wrap gap-2 pt-4",children:(0,n.jsx)(U.H,{field:{type:"relation",name:"users",label:"Users",relation:{schema:"user",field:"roles",type:"m2m",owner:!0,optional:!0}},fieldProps:{value:null!==(r=null==l?void 0:l.users)&&void 0!==r?r:[],onChange:o},content:null!=l?l:{}})})})})]})]})},Q=e=>{let{open:t,setOpen:r,resources:s,selectedPermissions:l,updateSelectedPermissions:a}=e;return(0,n.jsxs)(k.J2,{open:t,onOpenChange:r,children:[(0,n.jsx)(k.xo,{asChild:!0,children:(0,n.jsx)(y.z,{variant:"outline",className:"w-[150px] justify-start",children:"+ Add permission"})}),(0,n.jsx)(k.yk,{className:"p-0",align:"start",children:(0,n.jsxs)(z.mY,{children:[(0,n.jsx)(z.sZ,{placeholder:"Search for a permission"}),(0,n.jsxs)(z.e8,{children:[(0,n.jsx)(z.rb,{children:"No results found."}),(0,n.jsx)(z.fu,{children:(0,n.jsx)(W,{resources:null!=s?s:[],selectedPermissions:l,updateSelectedPermissions:a})})]})]})})]})},W=e=>{let{resources:t,selectedPermissions:r,updateSelectedPermissions:s}=e,l=t.filter(e=>!r.map(e=>e.resource).includes(e.id)),a=!l.length||t.every(e=>e.group);return(0,n.jsxs)(n.Fragment,{children:[l.map(e=>(0,n.jsx)(H,{resource:e,selectedPermissions:r,updateSelectedPermissions:s},e.id)),!a&&(0,n.jsx)(z.zz,{className:"my-2"})]})},H=e=>{var t;let{resource:r,selectedPermissions:s,updateSelectedPermissions:l}=e;return(null==r?void 0:null===(t=r.resources)||void 0===t?void 0:t.length)?(0,n.jsx)(W,{resources:r.resources,selectedPermissions:s,updateSelectedPermissions:l}):(0,n.jsx)(z.di,{value:r.id,onSelect:()=>{l([...s,{resource:r.id,value:"allow"}])},children:(0,F.Qs)(r.id)},r.id)},L=e=>{let{permission:t,resources:r,selectedPermissions:s,updateSelectedPermissions:l}=e,[i,o]=(0,a.useState)(!1),d=r.find(e=>e.id===t.resource);return console.log({resources:r,matchedResource:d,permission:t}),(0,n.jsxs)("div",{className:"themes-wrapper group relative flex flex-col overflow-hidden rounded-xl border shadow transition-all duration-200 ease-in-out hover:z-30",children:[(0,n.jsxs)("div",{className:"flex items-center gap-2 relative z-20 justify-end border-b bg-card px-3 py-2.5 text-card-foreground",children:[(0,n.jsxs)("div",{className:"flex items-center gap-1.5 pl-1 text-[13px] font-medium text-muted-foreground [&>svg]:h-[0.9rem] [&>svg]:w-[0.9rem]",children:[(0,n.jsx)(I.Z,{})," ",(0,F.Qs)(t.resource)]}),(0,n.jsxs)("div",{className:"ml-auto flex items-center gap-2 [&>form]:flex",children:[(0,n.jsxs)(V.u,{children:[(0,n.jsx)(V.aJ,{asChild:!0,children:(0,n.jsxs)(y.z,{size:"icon",variant:"outline",className:"[&_svg]-h-3.5 [&_svg]-h-3 h-6 w-6 rounded-[6px] bg-transparent text-foreground shadow-none hover:bg-muted dark:text-foreground [&_svg]:w-3",onClick:()=>{confirm("Are you sure?")&&l(s.filter(e=>e.resource!==t.resource))},children:[(0,n.jsx)("span",{className:"sr-only",children:"Remove"}),(0,n.jsx)(J.Z,{})]})}),(0,n.jsx)(V._v,{children:(0,n.jsx)("p",{children:"Remove permission"})})]}),(0,n.jsx)(Z.Z,{orientation:"vertical",className:"mx-0 hidden h-4 md:flex"}),(0,n.jsxs)(V.u,{children:[(0,n.jsx)(V.aJ,{asChild:!0,children:(0,n.jsxs)(y.z,{size:"icon",variant:"outline",className:"h-6 rounded-[6px] border bg-transparent px-2 text-xs text-foreground shadow-none hover:bg-muted dark:text-foreground",onClick:()=>o(!i),children:[(0,n.jsx)("span",{className:"sr-only",children:"Settings"}),(0,n.jsx)(A.Z,{})]})}),(0,n.jsx)(V._v,{children:(0,n.jsx)("p",{children:"Settings"})})]})]})]}),i&&(0,n.jsx)("div",{className:"relative z-10 [&>div]:rounded-none [&>div]:border-none [&>div]:shadow-none px-4 py-2.5 pb-4",children:(0,n.jsxs)(p.xJ,{children:[(0,n.jsx)(p.lX,{children:"Custom rule"}),(0,n.jsxs)(p.pf,{children:["Define a custom rule for this permission. For more information, see"," ",(0,n.jsx)("a",{target:"_blank",title:"Access Control",className:"underline text-primary",href:"https://fastschema.com/docs/reference/access-control",children:"Access Control"}),"."]}),(0,n.jsx)(p.NI,{children:(0,n.jsx)(X,{placeholder:"// Gets the current authenticated user ID\nlet authUserId = $context.User().ID;\n\n// Modify the request filter to include the author_id\nlet requestFilter = $context.Arg('filter', '{}');\nlet authorFilter = $sprintf('{\"author_id\": %d}', authUserId);\nlet combinedFilter = $sprintf('{\"$and\": [%s, %s]}', requestFilter, authorFilter);\n// Update the filter argument\n// Any function call must be assigned to a variable\nlet _ = $context.SetArg('filter', combinedFilter);\n\n// The last line is the return value, it should be a boolean\n// Check if the user has enough credit\nlet users = $db.Query($context, \"SELECT * FROM users WHERE id = ?\", authUserId);\nusers[0].Get(\"credit\") > 10",value:"allow"!==t.value?t.value:"",onChange:e=>{t.value=e,l(s.map(e=>e.resource===t.resource?t:e))}})}),(0,n.jsx)(p.zG,{})]})})]})},M=e=>{let{editingRole:t}=e,r=(0,i.useRouter)(),s=(0,u.cI)({resolver:(0,c.F)(x),defaultValues:null!=t?t:h,mode:"onChange"}),l=async e=>{try{let n=await (0,$.x$)(e,null==t?void 0:t.id);w.h4.success("Role ".concat(n.name," saved successfully.")),t||r.push("/settings/roles/edit?id=".concat(n.id)),s.reset(n)}catch(e){}};return(0,n.jsx)("div",{children:(0,n.jsx)(p.l0,{...s,children:(0,n.jsxs)("form",{className:"space-y-8",onSubmit:s.handleSubmit(l),children:[(0,n.jsx)(p.Wi,{control:s.control,name:"name",render:e=>{let{field:t}=e;return(0,n.jsxs)(p.xJ,{children:[(0,n.jsx)(p.lX,{className:"flex",children:(0,n.jsx)(g.u,{tip:"This is the name of your role.",icon:!0,children:(0,n.jsx)("span",{className:"mr-1",children:"Name"})})}),(0,n.jsx)(p.NI,{children:(0,n.jsx)(v.I,{...t,autoComplete:"auto",placeholder:"Role name"})}),(0,n.jsx)(p.zG,{})]})}}),(0,n.jsx)(p.Wi,{control:s.control,name:"description",render:e=>{let{field:t}=e;return(0,n.jsxs)(p.xJ,{children:[(0,n.jsx)(p.lX,{className:"flex",children:(0,n.jsx)("span",{className:"mr-1",children:"Description"})}),(0,n.jsx)(p.NI,{children:(0,n.jsx)(j.g,{...t,autoComplete:"auto",placeholder:"Role Description"})}),(0,n.jsx)(p.zG,{})]})}}),(0,n.jsx)(p.Wi,{control:s.control,name:"root",render:e=>{let{field:t}=e;return(0,n.jsx)(p.xJ,{children:(0,n.jsx)(p.NI,{children:(0,n.jsxs)("div",{className:"flex items-center space-x-2",children:[(0,n.jsx)(b.r,{checked:t.value,onCheckedChange:t.onChange,name:"root",id:"root","aria-readonly":!0}),(0,n.jsx)(N._,{htmlFor:"root",className:"flex align-middle gap-1",children:(0,n.jsx)(g.u,{tip:"Root roles have full access to all resources.",icon:!0,children:"Is root?"})})]})})})}}),(0,n.jsx)(p.Wi,{control:s.control,name:"rule",render:e=>{let{field:t}=e;return(0,n.jsxs)(p.xJ,{children:[(0,n.jsx)(p.lX,{className:"flex",children:(0,n.jsx)(g.u,{icon:!0,tip:"Rule to apply to all resources accessible by this role.",children:(0,n.jsx)("span",{className:"mr-1",children:"Global Rule"})})}),(0,n.jsx)(p.NI,{children:(0,n.jsx)(X,{value:t.value,onChange:t.onChange,placeholder:"$context.IP() in ['127.0.0.1']",height:"100px"})})]})}}),(0,n.jsx)(T,{editingRole:t,onPermissionsUpdated:e=>{s.setValue("permissions",e)},onUsersUpdated:e=>{(null==e?void 0:e.$add)&&s.setValue("$add.users",e.$add),(null==e?void 0:e.$clear)&&s.setValue("$clear.users",e.$clear)}}),(0,n.jsx)(y.z,{type:"submit",children:"Save"})]})})})};function K(){let e=(0,i.useSearchParams)().get("id"),{data:t,isLoading:r,error:c}=(0,s.a)({queryKey:["role",e],queryFn:()=>(0,$.cY)(e),retry:!1,refetchOnWindowFocus:!1});return((0,a.useEffect)(()=>((0,l.D8)({title:(null==t?void 0:t.name)?"Edit role: ".concat(t.name):"Create new role",description:(null==t?void 0:t.name)?"Edit role ".concat(t.name," to change the permissions of your users."):"Create a new role for your users.",breadcrumbs:[{name:"Roles",path:"/settings/roles"},{name:e?"Edit role":"New role",path:"/settings/roles/edit?id="+(null!=e?e:"")}]}),l.D8),[e,t]),r)?(0,n.jsx)(o.g,{}):c?(0,n.jsx)(d.T,{error:c}):(0,n.jsx)(M,{editingRole:t})}},15e3:function(e,t,r){"use strict";r.d(t,{T:function(){return o}});var n=r(57437),s=r(29069),l=r(62985),a=r(95601),i=r.n(a);let o=e=>{var t;let{title:r,description:a,error:o}=e;return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsx)(i(),{children:(0,n.jsx)("title",{children:null!=r?r:"Error"})}),(0,n.jsxs)(s.bZ,{variant:"destructive",children:[(0,n.jsx)(l.Z,{size:16,className:"mr-2"}),(0,n.jsx)(s.Cd,{children:null!=r?r:"Error"}),(null==o?void 0:o.message)||a&&(0,n.jsx)(s.X,{children:null!==(t=null==o?void 0:o.message)&&void 0!==t?t:a})]})]})}},75006:function(e,t,r){"use strict";r.d(t,{g:function(){return o}});var n=r(57437),s=r(22169),l=r(2252),a=r(95601),i=r.n(a);let o=e=>{let{title:t,description:r,error:a,full:o,className:d}=e;a&&(t="Error",r=a);let c=(0,s.cn)("z-50 bg-slate-100 opacity-75 flex flex-col items-center justify-center rounded-lg",null!=d?d:"",o?"fixed top-0 left-0 right-0 bottom-0 w-full h-screen":"w-full h-full");return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsx)(i(),{children:(0,n.jsx)("title",{children:null!=t?t:"Loading..."})}),(0,n.jsxs)("div",{className:c,children:[a?(0,n.jsx)(l.Z,{size:64,color:"#dc2626"}):(0,n.jsx)("div",{className:"loader ease-linear rounded-full border-4 border-t-4 border-gray-100 h-12 w-12 mb-4"}),(0,n.jsx)("h2",{className:"text-center text-black text-xl font-semibold",children:null!=t?t:"Loading..."}),(0,n.jsx)("p",{className:"w-1/3 text-center text-black",children:null!=r?r:"This may take a few seconds, please don't close this page."})]})]})}},82628:function(e,t,r){"use strict";r.d(t,{u:function(){return a}});var n=r(57437),s=r(62677),l=r(77252);let a=e=>{let{children:t,tip:r,icon:a}=e;return r?(0,n.jsxs)(s.u,{children:[(0,n.jsx)(s.aJ,{type:"button",children:(0,n.jsxs)("div",{className:"flex items-center gap-1",children:[t,a&&(0,n.jsx)(l.Z,{size:14})]})}),(0,n.jsx)(s._v,{sideOffset:10,children:r})]}):(0,n.jsx)(n.Fragment,{children:t})}},18157:function(e,t,r){"use strict";r.d(t,{EC:function(){return s},l:function(){return l},mb:function(){return a}});var n=r(31548);let s=e=>(null==e?void 0:e.type)==="m2m"||(null==e?void 0:e.type)==="o2m"&&!!(null==e?void 0:e.owner),l=(e,t,r,n,s,l)=>{if(!n){if(e.length){r(e[0]);return}if(!e.length&&t.length){r(s?null:void 0);return}r(s?null:void 0);return}let a=e.map(e=>({id:e.id})),i=a.length?a:void 0,o=t.map(e=>({id:e.id})),d=o.length?o:void 0,c=(null==i?void 0:i.length)||!s?i:null;r(l?{$add:i,$clear:d}:c)},a=async(e,t,r,s,l)=>{if(!s||!e.relation)return;let a={};if(t&&(a[s.label_field]={$like:"%".concat(t,"%")}),l){let t=e.relation.field;a["".concat(t,".id")]=l}return await (0,n.Q7)(s.name,{page:r,limit:20,filter:Object.keys(a).length?a:void 0,select:"id,".concat(s.label_field)})}},29069:function(e,t,r){"use strict";r.d(t,{Cd:function(){return d},X:function(){return c},bZ:function(){return o}});var n=r(57437),s=r(2265),l=r(57742),a=r(22169);let i=(0,l.j)("relative w-full rounded-lg border px-4 py-3 text-sm [&>svg+div]:translate-y-[-3px] [&>svg]:absolute [&>svg]:left-4 [&>svg]:top-4 [&>svg]:text-foreground [&>svg~*]:pl-7",{variants:{variant:{default:"bg-background text-foreground",destructive:"border-destructive/50 text-destructive dark:border-destructive [&>svg]:text-destructive"}},defaultVariants:{variant:"default"}}),o=s.forwardRef((e,t)=>{let{className:r,variant:s,...l}=e;return(0,n.jsx)("div",{ref:t,role:"alert",className:(0,a.cn)(i({variant:s}),r),...l})});o.displayName="Alert";let d=s.forwardRef((e,t)=>{let{className:r,...s}=e;return(0,n.jsx)("h5",{ref:t,className:(0,a.cn)("mb-1 font-medium leading-none tracking-tight",r),...s})});d.displayName="AlertTitle";let c=s.forwardRef((e,t)=>{let{className:r,...s}=e;return(0,n.jsx)("div",{ref:t,className:(0,a.cn)("text-sm [&_p]:leading-relaxed",r),...s})});c.displayName="AlertDescription"},9208:function(e,t,r){"use strict";r.d(t,{$N:function(){return f},Vq:function(){return o},cZ:function(){return u},fK:function(){return m}});var n=r(57437),s=r(2265),l=r(72936),a=r(62177),i=r(22169);let o=l.fC;l.xz;let d=l.h_;l.x8;let c=s.forwardRef((e,t)=>{let{className:r,...s}=e;return(0,n.jsx)(l.aV,{ref:t,className:(0,i.cn)("fixed inset-0 z-50 bg-black/80  data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",r),...s})});c.displayName=l.aV.displayName;let u=s.forwardRef((e,t)=>{let{className:r,children:s,...o}=e;return(0,n.jsxs)(d,{children:[(0,n.jsx)(c,{}),(0,n.jsxs)(l.VY,{ref:t,className:(0,i.cn)("fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4 border bg-background p-6 shadow-lg duration-200 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[state=closed]:slide-out-to-left-1/2 data-[state=closed]:slide-out-to-top-[48%] data-[state=open]:slide-in-from-left-1/2 data-[state=open]:slide-in-from-top-[48%] sm:rounded-lg",r),...o,children:[s,(0,n.jsxs)(l.x8,{className:"absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none data-[state=open]:bg-accent data-[state=open]:text-muted-foreground",children:[(0,n.jsx)(a.Pxu,{className:"h-4 w-4"}),(0,n.jsx)("span",{className:"sr-only",children:"Close"})]})]})]})});u.displayName=l.VY.displayName;let m=e=>{let{className:t,...r}=e;return(0,n.jsx)("div",{className:(0,i.cn)("flex flex-col space-y-1.5 text-center sm:text-left",t),...r})};m.displayName="DialogHeader";let f=s.forwardRef((e,t)=>{let{className:r,...s}=e;return(0,n.jsx)(l.Dx,{ref:t,className:(0,i.cn)("text-lg font-semibold leading-none tracking-tight",r),...s})});f.displayName=l.Dx.displayName,s.forwardRef((e,t)=>{let{className:r,...s}=e;return(0,n.jsx)(l.dk,{ref:t,className:(0,i.cn)("text-sm text-muted-foreground",r),...s})}).displayName=l.dk.displayName},82012:function(e,t,r){"use strict";r.d(t,{J2:function(){return i},xo:function(){return o},yk:function(){return d}});var n=r(57437),s=r(2265),l=r(57427),a=r(22169);let i=l.fC,o=l.xz;l.ee;let d=s.forwardRef((e,t)=>{let{className:r,align:s="center",sideOffset:i=4,...o}=e;return(0,n.jsx)(l.h_,{children:(0,n.jsx)(l.VY,{ref:t,align:s,sideOffset:i,className:(0,a.cn)("z-50 w-72 rounded-md border bg-popover p-4 text-popover-foreground shadow-md outline-none data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2",r),...o})})});d.displayName=l.VY.displayName},31548:function(e,t,r){"use strict";r.d(t,{Q7:function(){return i},Vj:function(){return l},il:function(){return o},rP:function(){return a},yd:function(){return d}});var n=r(18157),s=r(84971);let l=async(e,t,r)=>{let l={$add:{},$set:{},$clear:{}};for(let s of e.fields)if(s.name in t||s.is_system_field){if("relation"===s.type||"file"===s.type){var a,i;if(r){if(null===t[s.name]){l.$clear[s.name]=!0;continue}if(t[s.name].$nochange)continue;if(!(0,n.EC)(s.relation)){l.$set[s.name]=t[s.name];continue}t[s.name].$add&&(l.$add[s.name]=t[s.name].$add),t[s.name].$clear&&(l.$clear[s.name]=t[s.name].$clear);continue}if(null==t?void 0:null===(a=t[s.name])||void 0===a?void 0:a.$add){l[s.name]=t[s.name].$add;continue}if(null==t?void 0:null===(i=t[s.name])||void 0===i?void 0:i.$clear){delete l[s.name];continue}}l[s.name]=t[s.name]}return 0===Object.keys(l.$add).length&&delete l.$add,0===Object.keys(l.$set).length&&delete l.$set,0===Object.keys(l.$clear).length&&delete l.$clear,r?await (0,s.qb)("/content/".concat(e.name,"/").concat(r),l):await (0,s.SO)("/content/".concat(e.name),l)},a=e=>{let t={};return(null==e?void 0:e.limit)&&(t.limit=e.limit),(null==e?void 0:e.page)&&(t.page=e.page),(null==e?void 0:e.sort)&&(t.sort=e.sort),(null==e?void 0:e.select)&&(t.select=e.select),(null==e?void 0:e.filter)&&(t.filter=JSON.stringify(e.filter)),new URLSearchParams(t).toString()},i=async(e,t)=>{if(!e)throw Error("Schema name is required");let r={};(null==t?void 0:t.limit)&&(r.limit=t.limit),(null==t?void 0:t.page)&&(r.page=t.page),(null==t?void 0:t.sort)&&(r.sort=t.sort),(null==t?void 0:t.select)&&(r.select=t.select),(null==t?void 0:t.filter)&&(r.filter=JSON.stringify(t.filter));let n=Object.keys(r).length>0?"?"+a(t):"";return(0,s.dX)("/content/".concat(e).concat(n))},o=async(e,t,r)=>{let n=(null==r?void 0:r.length)?"?select=".concat(r.join(",")):"";return(0,s.dX)("/content/".concat(e.name,"/").concat(t).concat(n))},d=async(e,t)=>(0,s.HG)("/content/".concat(e,"/").concat(t))},97081:function(e,t,r){"use strict";r.d(t,{Qs:function(){return s},_v:function(){return n},rV:function(){return l}});let n=e=>new Promise(t=>setTimeout(t,e)),s=e=>{let t=e.toLowerCase().split(" ");for(let e=0;e<t.length;e++)t[e]=t[e].charAt(0).toUpperCase()+t[e].substring(1);return t.join(" ")},l=e=>s(e.replace(/[-_]/g," "))},90688:function(e,t,r){"use strict";r.d(t,{Rd:function(){return i},cY:function(){return l},x$:function(){return a},xv:function(){return s}});var n=r(84971);let s=async()=>(0,n.dX)("/role"),l=async e=>{var t;if(!e)return null;let r=await (0,n.dX)("/role/".concat(e));return r.permissions=null!==(t=r.permissions)&&void 0!==t?t:[],r},a=async(e,t)=>(t||(delete e.$add,delete e.$clear),t?await (0,n.qb)("/role/".concat(t),e):await (0,n.SO)("/role",e)),i=e=>(0,n.HG)("/role/".concat(e))}},function(e){e.O(0,[310,401,637,789,792,829,32,396,160,147,340,334,862,971,69,744],function(){return e(e.s=73421)}),_N_E=e.O()}]);