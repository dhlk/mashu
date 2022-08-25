import bpy
from bpy_extras import image_utils

def findAttach(material):
	for node in material.node_tree.nodes:
		if node.label == "attach":
				return node
	return None

def attachVideo(name, path):
	bpy.data.speakers[name].sound = bpy.data.sounds.load(filepath=path)
	material = bpy.data.materials[name]
	attach = findAttach(material)
	attach.image = image_utils.load_image(imagepath=path)
	_ = attach.image.frame_duration # dumb bug; first read returns 1; read twice
	attach.image_user.frame_duration = attach.image.frame_duration
	attach.image_user.frame_start = material.get("frame_start", 1)

def setup(args):
	bpy.context.scene.render.engine = 'BLENDER_EEVEE'
	bpy.context.scene.render.ffmpeg.audio_bitrate = args.bitrate
	bpy.context.scene.render.ffmpeg.audio_codec = args.acodec
	bpy.context.scene.render.ffmpeg.audio_mixrate = args.samplerate
	bpy.context.scene.render.ffmpeg.codec = args.vcodec
	bpy.context.scene.render.ffmpeg.constant_rate_factor = args.quality
	bpy.context.scene.render.ffmpeg.ffmpeg_preset = args.speed
	bpy.context.scene.render.ffmpeg.format = args.format
	bpy.context.scene.render.ffmpeg.gopsize = args.gopsize
	bpy.context.scene.render.filepath = args.output
	bpy.context.scene.render.fps = args.fps
	bpy.context.scene.render.image_settings.color_mode = "RGB"
	bpy.context.scene.render.image_settings.file_format = "FFMPEG"
	bpy.context.scene.render.resolution_x = args.width
	bpy.context.scene.render.resolution_y = args.height
	bpy.context.scene.render.use_stamp_camera = False
	bpy.context.scene.render.use_stamp_date = False
	bpy.context.scene.render.use_stamp_filename = False
	bpy.context.scene.render.use_stamp_frame = False
	bpy.context.scene.render.use_stamp_frame_range = False
	bpy.context.scene.render.use_stamp_hostname = False
	bpy.context.scene.render.use_stamp_lens = False
	bpy.context.scene.render.use_stamp_marker = False
	bpy.context.scene.render.use_stamp_memory = False
	bpy.context.scene.render.use_stamp_render_time = False
	bpy.context.scene.render.use_stamp_scene = False
	bpy.context.scene.render.use_stamp_sequencer_strip = False
	bpy.context.scene.render.use_stamp_time = False
	bpy.context.scene.eevee.taa_render_samples = args.samples

	for attach in args.attach:
		attachVideo(*attach)

# object to attach sound to, material to project on, path to video
def main():
	import sys
	import argparse

	parser = argparse.ArgumentParser(description="Process a mashu blender render.")
	parser.add_argument("-acodec", required=True)
	parser.add_argument("-bitrate", type=int, required=True, help="Audio bitrate (kb/s)")
	parser.add_argument("-format", required=True)
	parser.add_argument("-fps", type=int, required=True)
	parser.add_argument("-gopsize", type=int, required=True)
	parser.add_argument("-height", type=int, required=True)
	parser.add_argument("-output", required=True)
	parser.add_argument("-quality", required=True, help="Video output quality")
	parser.add_argument("-samplerate", type=int, required=True, help="Audio samplerate (hz)")
	parser.add_argument("-samples", type=int, required=True, help="Video render samples")
	parser.add_argument("-speed", required=True, help="Video encoding speed")
	parser.add_argument("-vcodec", required=True)
	parser.add_argument("-width", type=int, required=True)
	parser.add_argument("-attach", action="append", nargs=2, help="Attach audio/video to an speaker/material (-attach name filepath)")

	argv = sys.argv
	if "--" not in argv:
		argv = []
	else:
		argv = argv[argv.index("--") + 1:]

	args = parser.parse_args(argv)
	try:
		setup(args)
	except Exception as ex:
		import traceback
		traceback.print_exception(type(ex), ex, ex.__traceback__)
		raise SystemExit(1)

if __name__ == "__main__":
	main()

# TODO support/test multi-scene
# TODO support text (set value, font size, font)
# TODO validate output does not exist or is empty
