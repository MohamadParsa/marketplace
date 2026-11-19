from pathlib import Path
from PIL import Image,ImageDraw,ImageFont
root=Path.cwd();out=root/'output'/'playwright';out.mkdir(exist_ok=True)
font=ImageFont.truetype('C:/Windows/Fonts/segoeui.ttf',24)
small=ImageFont.truetype('C:/Windows/Fonts/segoeui.ttf',16)
def board(name,files,labels,crop):
    tilew=560;tileh=550;pad=20
    result=Image.new('RGB',(3*tilew+4*pad,tileh+120),'#f7f2f4');d=ImageDraw.Draw(result)
    d.text((24,18),'CHERRIVA  |  DESIGN REVIEW  |  2026-11-17',font=font,fill='#88163c')
    for i,(p,label) in enumerate(zip(files,labels)):
        im=Image.open(root/p).convert('RGB');im=im.crop((0,0,im.width,min(im.height,crop)));im.thumbnail((tilew,tileh))
        x=pad+i*(tilew+pad);d.text((x,64),label,font=small,fill='#3d2732');result.paste(im,(x,102));d.rectangle((x-1,101,x+im.width,102+im.height),outline='#d8bdc9')
    result.save(out/name)
board('storefront-three-directions.png',[f'01-storefront/images/revision-v0-4/home-{x}-1440.png' for x in ['balanced','sales','editorial']],['A  Balanced & elegant','B  Sales focused','C  Luxury & editorial'],1400)
board('backoffice-three-directions.png',[f'02-backoffice/output/playwright/bo-01-dashboard-{x}.png' for x in ['a','b','c']],['A  Balanced & elegant','B  Sales focused','C  Luxury & editorial'],1400)
print('Created 2 comparison boards')
