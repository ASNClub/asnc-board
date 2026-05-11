import { useState, useCallback } from 'react';
import Cropper from 'react-easy-crop';

const createCroppedBlob = (imageSrc, crop, fileName, minWidth = 0) =>
  new Promise((resolve) => {
    const img = new Image();
    img.onload = () => {
      const scale = minWidth > 0 && crop.width < minWidth ? minWidth / crop.width : 1;
      const canvas = document.createElement('canvas');
      canvas.width = Math.round(crop.width * scale);
      canvas.height = Math.round(crop.height * scale);
      const ctx = canvas.getContext('2d');
      ctx.drawImage(img, crop.x, crop.y, crop.width, crop.height, 0, 0, canvas.width, canvas.height);
      canvas.toBlob((blob) => {
        resolve(new File([blob], fileName, { type: 'image/jpeg' }));
      }, 'image/jpeg', 0.92);
    };
    img.src = imageSrc;
  });

const ImageCropModal = ({ image, aspect, hint, minWidth = 0, onDone, onCancel }) => {
  const [crop, setCrop] = useState({ x: 0, y: 0 });
  const [zoom, setZoom] = useState(1);
  const [croppedArea, setCroppedArea] = useState(null);
  const [busy, setBusy] = useState(false);

  const onCropComplete = useCallback((_area, pixels) => {
    setCroppedArea(pixels);
  }, []);

  const handleSave = async () => {
    if (!croppedArea) return;
    setBusy(true);
    const file = await createCroppedBlob(image, croppedArea, 'cropped.jpg', minWidth);
    onDone(file);
  };

  return (
    <div className="modal-overlay" onClick={onCancel}>
      <div className="modal" onClick={(e) => e.stopPropagation()} style={{ maxWidth: 560, padding: 0 }}>
        <div style={{ position: 'relative', width: '100%', height: 360, background: '#1a1a1a' }}>
          <Cropper
            image={image}
            crop={crop}
            zoom={zoom}
            aspect={aspect}
            onCropChange={setCrop}
            onZoomChange={setZoom}
            onCropComplete={onCropComplete}
          />
        </div>
        <div style={{ padding: '12px 16px', display: 'flex', alignItems: 'center', gap: 12 }}>
          <input
            type="range"
            min={1}
            max={3}
            step={0.05}
            value={zoom}
            onChange={(e) => setZoom(Number(e.target.value))}
            style={{ flex: 1 }}
          />
          {hint && <span style={{ fontFamily: 'var(--mono)', fontSize: 10.5, color: 'var(--text-dim)', whiteSpace: 'nowrap' }}>{hint}</span>}
        </div>
        <div style={{ padding: '0 16px 14px', display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
          <button className="btn ghost" onClick={onCancel}>отмена</button>
          <button className="btn primary" onClick={handleSave} disabled={busy}>
            {busy ? 'обрезаем…' : 'применить'}
          </button>
        </div>
      </div>
    </div>
  );
};

export default ImageCropModal;
