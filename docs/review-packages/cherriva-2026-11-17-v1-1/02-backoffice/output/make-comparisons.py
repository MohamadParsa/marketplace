from pathlib import Path
from PIL import Image, ImageDraw, ImageFont

root=Path(__file__).resolve().parent / 'playwright'
font=ImageFont.truetype('C:/Windows/Fonts/arial.ttf',22)
small=ImageFont.truetype('C:/Windows/Fonts/arial.ttf',14)
for name,source in [('dashboard','bo-01-dashboard'),('login','login')]:
    images=[Image.open(root/f'{source}-{theme}.png').convert('RGB') for theme in 'abc']
    thumb_w=480
    heights=[round(img.height*thumb_w/img.width) for img in images]
    canvas=Image.new('RGB',(1504,max(heights)+96),'#f5f2f4')
    draw=ImageDraw.Draw(canvas)
    for i,(image,height,label) in enumerate(zip(images,heights,['A / Balanced','B / Sales-focused','C / Luxury'])):
        x=16+i*496
        draw.rounded_rectangle((x,16,x+480,72),radius=9,fill='white')
        draw.text((x+16,27),label,font=font,fill='#7c1937')
        draw.text((x+16,53),'Cherriva / Backoffice design v0.1',font=small,fill='#8a7180')
        canvas.paste(image.resize((thumb_w,height),Image.Resampling.LANCZOS),(x,80))
    canvas.save(root/f'comparison-{name}.png')
print('Created two comparison sheets')
